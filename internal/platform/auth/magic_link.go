package auth

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"

	"github.com/gomodule/redigo/redis"
	"golang.org/x/crypto/bcrypt"
)

type UserInfo struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	AccountID   string `json:"account_id"`
	AccountName string `json:"account_name"`
	DraftID     string `json:"draft_id"`
	NewUser     bool   `json:"new_user"`
	MemberID    string `json:"member_id"`
}

func (usrInfo *UserInfo) encode() []byte {
	json, err := json.Marshal(usrInfo)
	if err != nil {
		log.Println("***> unexpected/unhandled error in magic link when marshaling message. error:", err)
	}
	return json
}

func (usrInfo *UserInfo) UnmarshalJSON(data []byte) error {

	type Alias UserInfo
	usInfo := &struct {
		*Alias
	}{
		Alias: (*Alias)(usrInfo),
	}
	if err := json.Unmarshal(data, &usInfo); err != nil {
		return err
	}
	return nil
}

func GenerateRandomToken(n int) (string, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), err
}

func CreateMagicLink(accountID, name, emailAddress, memId string, rp *redis.Pool) (string, error) {
	token, err := GenerateRandomToken(32)
	if err != nil {
		return "", err
	}

	userInfo := UserInfo{
		Name:      name,
		AccountID: accountID,
		Email:     emailAddress,
		MemberID:  memId,
	}

	log.Printf("userInfo --- %+v", userInfo)

	err = setToken(token, userInfo, rp)
	if err != nil {
		return "", err
	}

	magicLink := fmt.Sprintf("https://baserelay.com/home/join?token=%v", token)

	log.Println("join magicLink-------> ", magicLink)

	return magicLink, nil
}

func CreateMagicLaunchLink(draftID, accountName, emailAddress string, rp *redis.Pool) (string, error) {
	token, err := GenerateRandomToken(32)
	if err != nil {
		return "", err
	}

	userInfo := UserInfo{
		AccountName: accountName,
		DraftID:     draftID,
		Email:       emailAddress,
	}

	err = setToken(token, userInfo, rp)
	if err != nil {
		return "", err
	}

	magicLink := fmt.Sprintf("https://baserelay.com/home/launch?token=%v", token)

	log.Println("launch magicLink-------> ", magicLink)

	return magicLink, nil
}

func EmailHash(emailAddress string) (string, error) {
	bmHash, err := bcrypt.GenerateFromPassword([]byte(emailAddress), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(bmHash), nil
}

// Invoke this when the user hits
// the login URL https://bookface.com/login?code=<token>
func AuthenticateToken(token string, rp *redis.Pool) (UserInfo, error) {
	usrInfo, err := getUserInfo(token, rp)
	if err != nil {
		return UserInfo{}, err
	}
	// Done! The user with username emailAddress is now authenticated!
	return usrInfo, nil
}

func setToken(key string, usrInfo UserInfo, rp *redis.Pool) error {
	conn := rp.Get()
	defer conn.Close()
	err := conn.Send("SET", key, usrInfo.encode())
	return err
}

func getUserInfo(key string, rp *redis.Pool) (UserInfo, error) {
	conn := rp.Get()
	defer conn.Close()
	msgStr, err := redis.String(conn.Do("GET", key))
	if err != nil {
		return UserInfo{}, err
	}

	userInfo := &UserInfo{}
	userInfo.UnmarshalJSON([]byte(msgStr))
	return *userInfo, nil
}
