package job

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
)

const (
	reminders = "zset:reminders"
	delay     = "zset:delay"
)

type RedisJob struct {
	AccountID string
	EntityID  string
	ItemID    string
	Time      int64
	State     int
}

const (
	JobStateQueud = 0
	JobStateRiped = 1
)

type Listener struct {
}

func (l Listener) AddReminder(accountID, entityID, itemID string, when time.Time, rp *redis.Pool) error {
	conn := rp.Get()
	defer conn.Close()
	whenMilli := util.GetMilliSeconds(when)
	rrj := RedisJob{
		AccountID: accountID,
		EntityID:  entityID,
		ItemID:    itemID,
		Time:      whenMilli,
	}

	raw, err := json.Marshal(rrj)
	if err != nil {
		return err
	}

	return zadd(conn, reminders, whenMilli, string(raw))
}

func (l Listener) RunReminderListener(db *sqlx.DB, rp *redis.Pool) {
	conn := rp.Get()
	defer conn.Close()

	for {
		//the locks are approximate. check the item state before proceding with the operation. (Two clients should not execute the next node/send push notifications)
		redisJob, err := zpop(conn, reminders)
		if err != nil && err != redis.ErrNil {
			log.Println("unexpected error in RunListener. err: ", err)
		}
		if redisJob.State == JobStateRiped {
			//do the notifications
			log.Printf("perform notifications.... %+v \n", redisJob)
			(&Job{}).EventItemReminded(redisJob.AccountID, redisJob.EntityID, redisJob.ItemID, db, rp)
		}
		time.Sleep(3 * time.Second) //reduce this time when more requests received
	}
}

func zpop(c redis.Conn, key string) (result RedisJob, err error) {
	var redisJob RedisJob
	defer func() {
		// Return connection to normal state on error.
		if err != nil {
			c.Do("DISCARD")
		}
	}()

	// Loop until transaction is successful.
	for {
		if _, err := c.Do("WATCH", key); err != nil {
			return redisJob, err
		}

		members, err := redis.Strings(c.Do("ZRANGE", key, 0, 0))
		if err != nil {
			return redisJob, err
		}
		if len(members) != 1 {
			return redisJob, redis.ErrNil
		}

		member := members[0]
		err = json.Unmarshal([]byte(member), &redisJob)
		if err != nil {
			return redisJob, err
		}

		//if not riped. Keep it.
		if redisJob.Time <= time.Now().Unix() {
			c.Send("MULTI")
			c.Send("ZREM", key, member)
			queued, err := c.Do("EXEC")
			if err != nil {
				return redisJob, err
			}
			if queued != nil {
				redisJob.State = JobStateRiped
				result = redisJob
				break
			}
		}

	}

	return result, nil
}

func zadd(c redis.Conn, key string, whenMilli int64, raw string) error {
	err := c.Send("ZADD", key, whenMilli, raw)
	if err != nil {
		return err
	}
	return nil
}
