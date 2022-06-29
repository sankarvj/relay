package job

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/account"
	"gitlab.com/vjsideprojects/relay/internal/notification"
	"gitlab.com/vjsideprojects/relay/internal/platform/stream"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
)

const (
	reminders = "zset:reminders"
	delay     = "zset:delay"
)

type RedisJob struct {
	AccountID string
	UserID    string
	EntityID  string
	ItemID    string
	Time      int64
	State     int
	Type      int
	Meta      map[string]interface{}
}

const (
	JobStateQueued = 0
	JobStateRiped  = 1
)

const (
	JobTypeReminder = 1
	JobTypeDelay    = 2
)

type Listener struct {
}

func (j *Job) AddDelay(accountID, userID, entityID, itemID string, meta map[string]interface{}, when time.Time, rp *redis.Pool) error {
	conn := rp.Get()
	defer conn.Close()
	whenMilli := util.GetMilliSeconds(when)
	rrj := RedisJob{
		AccountID: accountID,
		UserID:    userID,
		EntityID:  entityID,
		ItemID:    itemID,
		Time:      whenMilli,
		Meta:      meta,
		Type:      JobTypeDelay,
	}

	raw, err := json.Marshal(rrj)
	if err != nil {
		return err
	}

	return zadd(conn, reminders, whenMilli, string(raw))
}

func (j *Job) AddReminder(accountID, userID, entityID, itemID string, when time.Time, rp *redis.Pool) error {
	conn := rp.Get()
	defer conn.Close()
	whenMilli := util.GetMilliSeconds(when)
	rrj := RedisJob{
		AccountID: accountID,
		UserID:    userID,
		EntityID:  entityID,
		ItemID:    itemID,
		Time:      whenMilli,
		Type:      JobTypeReminder,
	}

	raw, err := json.Marshal(rrj)
	if err != nil {
		return err
	}

	return zadd(conn, reminders, whenMilli, string(raw))
}

func (J *Job) AddVisitor(accountID, visitorID, body string, db *sqlx.DB, rp *redis.Pool) error {
	log.Println("*> Reached addvisitor on job")
	err := notification.VisitorInvitation(accountID, visitorID, body, db, rp)
	return err
}

func (J *Job) AddMember(accountID, memberID, userName, userEmail, body string, db *sqlx.DB, rp *redis.Pool) error {
	ctx := context.Background()
	log.Println("*> Reached addmember on job")
	a, err := account.Retrieve(ctx, db, accountID)
	if err != nil {
		log.Println("***>***> VisitorInvitation: unexpected/unhandled error occurred when retriving account. error:", err)
		return err
	}
	requester := fmt.Sprintf("Admin from %s", a.Name)
	err = notification.JoinInvitation(accountID, a.Name, requester, userName, userEmail, memberID, db, rp)
	return err
}

func (l Listener) RunReminderListener(db *sqlx.DB, rp *redis.Pool, fbSDKPath string) {
	conn := rp.Get()
	defer conn.Close()

	for {
		log.Println("internalRunReminderListener: Listening...")
		//the locks are approximate. check the item state before proceding with the operation. (Two clients should not execute the next node/send push notifications)
		redisJob, err := zpop(conn, reminders)
		if err != nil && err != redis.ErrNil {
			log.Println("*> expected error in RunListener. No reminders/delay exist in queue. Ignore", err)
		}
		if redisJob.State == JobStateRiped {
			switch redisJob.Type {
			case JobTypeReminder:
				go NewJob(db, rp, fbSDKPath).Stream(stream.NewReminderMessage(redisJob.AccountID, redisJob.UserID, redisJob.EntityID, redisJob.ItemID))
			case JobTypeDelay:
				go NewJob(db, rp, fbSDKPath).Stream(stream.NewDelayMessage(redisJob.AccountID, redisJob.UserID, redisJob.EntityID, redisJob.ItemID, redisJob.Meta))
			}
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
		time.Sleep(3 * time.Second)
		if _, err := c.Do("WATCH", key); err != nil {
			return redisJob, err
		}

		members, err := redis.Strings(c.Do("ZRANGE", key, 0, 0))
		if err != nil {
			return redisJob, err
		}

		//if no members available
		if len(members) != 1 {
			return redisJob, redis.ErrNil
		}

		member := members[0]
		err = json.Unmarshal([]byte(member), &redisJob)
		if err != nil {
			return redisJob, err
		}

		//if riped. Remove it.
		now := util.GetMilliSeconds(time.Now())
		if redisJob.Time < now {
			c.Send("MULTI")
			c.Send("ZREM", key, member)
			_, err := c.Do("EXEC")
			if err != nil {
				return redisJob, err
			}
			log.Printf("internal.job.job_listener riped : type %d\n", redisJob.Type)
			redisJob.State = JobStateRiped
			result = redisJob
			break
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
