package auth

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"

	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/platform/database"
	"golang.org/x/crypto/bcrypt"
)

type UserInfo struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	AccountID   string `json:"account_id"`
	AccountName string `json:"account_name"`
	DraftID     string `json:"draft_id"`
	Verified    bool   `json:"verified"`
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

func SimpleLink(accountID, teamID, entityID, itemID string) string {
	magicLink := fmt.Sprintf("https://workbaseone.com/v1/accounts/%s/teams/%s/entities/%s/items/%s", accountID, teamID, entityID, itemID)
	return magicLink
}

func CreateMagicLink(workBaseDomain, accountID, name, emailAddress, memId string, sdb *database.SecDB) (string, error) {
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

	err = sdb.SetUserToken(token, userInfo.encode())
	if err != nil {
		return "", err
	}

	magicLink := fmt.Sprintf("https://%s/home/join?token=%v", workBaseDomain, token)

	log.Println("join magicLink-------> ", magicLink)

	return magicLink, nil
}

func CreateVisitorMagicLink(accountID, name, emailAddress, visitorID, token string, sdb *database.SecDB) (string, error) {

	userInfo := UserInfo{
		Name:      name,
		AccountID: accountID,
		Email:     emailAddress,
		MemberID:  visitorID,
	}

	err := sdb.SetUserToken(token, userInfo.encode())
	if err != nil {
		return "", err
	}

	magicLink := fmt.Sprintf("https://workbaseone.com/home/visit?token=%v", token)

	log.Println("join magicLink-------> ", magicLink)

	return magicLink, nil
}

func CreateMagicLaunchLink(domain, draftID, accountName, emailAddress string, sdb *database.SecDB) (string, error) {
	token, err := GenerateRandomToken(32)
	if err != nil {
		return "", err
	}

	userInfo := UserInfo{
		AccountName: accountName,
		DraftID:     draftID,
		Email:       emailAddress,
	}

	err = sdb.SetUserToken(token, userInfo.encode())
	if err != nil {
		return "", errors.Wrap(err, "redis connection error")
	}

	magicLink := fmt.Sprintf("https://%s/home/launch?token=%v", domain, token)

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
func AuthenticateToken(token string, sdb *database.SecDB) (UserInfo, error) {
	usrInfo, err := getUserInfo(token, sdb)
	if err != nil {
		return UserInfo{}, err
	}
	// Done! The user with the emailAddress is now authenticated!
	return usrInfo, nil
}

func getUserInfo(key string, sdb *database.SecDB) (UserInfo, error) {
	msgStr, err := sdb.GetUserToken(key)
	if err != nil {
		return UserInfo{}, err
	}
	userInfo := &UserInfo{}
	userInfo.UnmarshalJSON([]byte(msgStr))
	//TODO delete token from redis after the first time retrival
	return *userInfo, nil
}
