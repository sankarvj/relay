package user

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/platform/auth"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
	"go.opencensus.io/trace"
	"golang.org/x/crypto/bcrypt"
)

const usersCollection = "users"

var (
	// ErrNotFound is used when a specific User is requested but does not exist.
	ErrNotFound = errors.New("User not found")

	// ErrInvalidID occurs when an ID is not in a valid form.
	ErrInvalidID = errors.New("ID is not in its proper form")

	// ErrAuthenticationFailure occurs when a user attempts to authenticate but
	// anything goes wrong.
	ErrAuthenticationFailure = errors.New("Authentication failed")

	// ErrForbidden occurs when a user tries to do something that is forbidden to them according to our access control policies.
	ErrForbidden = errors.New("Attempted action is not allowed")
)

// List retrieves a list of existing users from the database.
func List(ctx context.Context, db *sqlx.DB) ([]User, error) {
	ctx, span := trace.StartSpan(ctx, "internal.user.List")
	defer span.End()

	users := []User{}
	const q = `SELECT * FROM users`

	if err := db.SelectContext(ctx, &users, q); err != nil {
		return nil, errors.Wrap(err, "selecting users")
	}

	return users, nil
}

// RetrieveCurrentUserID gets the current user wrt the auth claims from the database.
func RetrieveCurrentUserID(ctx context.Context) (string, error) {
	ctx, span := trace.StartSpan(ctx, "internal.user.RetrieveCurrentUserID")
	defer span.End()

	claims, ok := ctx.Value(auth.Key).(auth.Claims)
	if !ok {
		return "", ErrNotFound
	}
	return claims.Subject, nil
}

func RetrieveWSCurrentUserID(ctx context.Context) (string, error) {
	ctx, span := trace.StartSpan(ctx, "internal.user.RetrieveWSCurrentUserID")
	defer span.End()

	userID, ok := ctx.Value(auth.SocketKey).(string)
	if !ok {
		return "", ErrNotFound
	}
	return userID, nil
}

// Retrieve gets the specified user from the database.
func Retrieve(ctx context.Context, claims auth.Claims, db *sqlx.DB) (*User, error) {
	ctx, span := trace.StartSpan(ctx, "internal.user.Retrieve")
	defer span.End()

	id := claims.StandardClaims.Subject
	if _, err := uuid.Parse(id); err != nil {
		return nil, ErrInvalidID
	}

	// If you are not an admin and looking to retrieve someone else then you are rejected.
	if !claims.HasRole(auth.RoleAdmin) && claims.Subject != id {
		return nil, ErrForbidden
	}

	var u User
	const q = `SELECT * FROM users WHERE user_id = $1`
	if err := db.GetContext(ctx, &u, q, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}

		return nil, errors.Wrapf(err, "selecting user %q", id)
	}

	return &u, nil
}

func RetrieveCurrentUser(ctx context.Context, db *sqlx.DB) (*User, error) {
	ctx, span := trace.StartSpan(ctx, "internal.user.RetrieveCurrentUser")
	defer span.End()

	currentUserID, err := RetrieveCurrentUserID(ctx)
	if err != nil {
		return nil, err
	}

	return RetrieveUser(ctx, db, currentUserID)
}

func RetrieveUser(ctx context.Context, db *sqlx.DB, currentUserID string) (*User, error) {
	var u User
	const q = `SELECT * FROM users WHERE user_id = $1`
	if err := db.GetContext(ctx, &u, q, currentUserID); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, errors.Wrapf(err, "selecting current user %q", currentUserID)
	}

	return &u, nil
}

func RetrieveUserByUniqIdentifier(ctx context.Context, db *sqlx.DB, email, phone string) (User, error) {
	var u User
	const q = `select * from users where (email=$1 AND email != '') OR (phone=$2 AND phone != '')`
	if err := db.GetContext(ctx, &u, q, email, phone); err != nil {
		if err == sql.ErrNoRows {
			return User{}, ErrNotFound
		}
		return User{}, errors.Wrapf(err, "selecting current user by email: %s or phone: %s", email, phone)
	}

	return u, nil
}

func BulkRetrieveUsers(ctx context.Context, ids []string, db *sqlx.DB) ([]User, error) {
	ctx, span := trace.StartSpan(ctx, "internal.user.BulkRetrieveUsers")
	defer span.End()

	users := []User{}
	const q = `SELECT * FROM users where user_id = any($1)`

	if err := db.SelectContext(ctx, &users, q, pq.Array(ids)); err != nil {
		return users, errors.Wrap(err, "selecting bulk users for selected user ids")
	}

	return users, nil
}

// Create inserts a new user into the database.
func Create(ctx context.Context, db *sqlx.DB, n NewUser, now time.Time) (User, error) {
	ctx, span := trace.StartSpan(ctx, "internal.user.Create")
	defer span.End()

	hash, err := bcrypt.GenerateFromPassword([]byte(n.Password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, errors.Wrap(err, "generating password hash")
	}

	accountBytes, err := json.Marshal(n.Accounts)
	if err != nil {
		return User{}, errors.Wrap(err, "encode accounts to bytes")
	}

	u := User{
		ID:           uuid.New().String(),
		Accounts:     util.String(string(accountBytes)),
		Name:         &n.Name,
		Email:        n.Email,
		Avatar:       n.Avatar,
		Phone:        n.Phone,
		Provider:     n.Provider,
		Verified:     n.Verified,
		IssuedAt:     now.UTC(),
		Roles:        n.Roles,
		PasswordHash: hash,
		CreatedAt:    now.UTC(),
		UpdatedAt:    now.UTC().Unix(),
	}

	const q = `INSERT INTO users
		(user_id, accounts, name, email, avatar, phone, provider, verified, issued_at, password_hash, roles, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`
	_, err = db.ExecContext(
		ctx, q,
		u.ID, u.Accounts, u.Name, u.Email, u.Avatar, u.Phone, u.Provider, u.Verified, u.IssuedAt,
		u.PasswordHash, u.Roles,
		u.CreatedAt, u.UpdatedAt,
	)
	if err != nil {
		return User{}, errors.Wrap(err, "inserting user")
	}

	return u, nil
}

// Update replaces a user document in the database.
func Update(ctx context.Context, claims auth.Claims, db *sqlx.DB, id string, upd UpdateUser, now time.Time) error {
	ctx, span := trace.StartSpan(ctx, "internal.user.Update")
	defer span.End()

	u, err := Retrieve(ctx, claims, db)
	if err != nil {
		return err
	}

	if upd.Name != nil {
		u.Name = upd.Name
	}

	if upd.Phone != nil {
		u.Phone = upd.Phone
	}

	if upd.Avatar != nil {
		u.Avatar = upd.Avatar
	}

	if upd.Password != nil {
		pw, err := bcrypt.GenerateFromPassword([]byte(*upd.Password), bcrypt.DefaultCost)
		if err != nil {
			return errors.Wrap(err, "generating password hash")
		}
		u.PasswordHash = pw
	}

	u.UpdatedAt = now.Unix()

	const q = `UPDATE users SET
		"name" = $2,
		"email" = $3,
		"roles" = $4,
		"password_hash" = $5,
		"updated_at" = $6,
		"phone" = $7 
		WHERE user_id = $1`
	_, err = db.ExecContext(ctx, q, id,
		u.Name, u.Email, u.Roles,
		u.PasswordHash, u.UpdatedAt, u.Phone,
	)
	if err != nil {
		return errors.Wrap(err, "updating user")
	}

	return nil
}

func UpdatePassword(ctx context.Context, db *sqlx.DB, userID string, password string, now time.Time) error {
	//TODO exploitation possible. Anyone without claims can update the password it seems
	ctx, span := trace.StartSpan(ctx, "internal.user.UpdatePassword")
	defer span.End()

	pw, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return errors.Wrap(err, "generating password hash")
	}

	const q = `UPDATE users SET "password_hash" = $2, "updated_at" = $3
		WHERE user_id = $1`
	_, err = db.ExecContext(ctx, q, userID,
		pw, now.Unix(),
	)
	if err != nil {
		return errors.Wrap(err, "updating user password")
	}

	return nil
}

func RemoveIndex(s pq.StringArray, index int) pq.StringArray {
	return append(s[:index], s[index+1:]...)
}

func SetAsVerified(ctx context.Context, db *sqlx.DB, u *User, now time.Time) error {
	ctx, span := trace.StartSpan(ctx, "internal.user.SetAsVerified")
	defer span.End()

	u.Verified = true
	u.UpdatedAt = now.Unix()

	const q = `UPDATE users SET
		"verified" = $2,
		"updated_at" = $3
		WHERE user_id = $1`
	_, err := db.ExecContext(ctx, q, u.ID,
		u.Verified, u.UpdatedAt,
	)
	if err != nil {
		return errors.Wrap(err, "setting the user as verified")
	}

	return nil
}

// Delete removes a user from the database.
func Delete(ctx context.Context, db *sqlx.DB, id string) error {
	ctx, span := trace.StartSpan(ctx, "internal.user.Delete")
	defer span.End()

	if _, err := uuid.Parse(id); err != nil {
		return ErrInvalidID
	}

	const q = `DELETE FROM users WHERE user_id = $1`

	if _, err := db.ExecContext(ctx, q, id); err != nil {
		return errors.Wrapf(err, "deleting user %s", id)
	}

	return nil
}

func removeDuplicateValues(intSlice pq.StringArray) pq.StringArray {
	keys := make(map[interface{}]bool)
	list := pq.StringArray{}
	for _, entry := range intSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func (u *User) UpdateAccounts(ctx context.Context, db *sqlx.DB, accounts map[string]interface{}) error {
	ctx, span := trace.StartSpan(ctx, "internal.user.UpdateAccounts")
	defer span.End()

	input, err := json.Marshal(accounts)
	if err != nil {
		return errors.Wrap(err, "encode meta to input")
	}
	inputB := string(input)
	u.Accounts = &inputB

	const q = `UPDATE users SET
		"accounts" = $2 
		WHERE user_id = $1`
	_, err = db.ExecContext(ctx, q, u.ID,
		u.Accounts,
	)
	return err
}

func (u User) AccountsB() map[string]interface{} {
	accounts := make(map[string]interface{}, 0)
	if u.Accounts == nil || *u.Accounts == "" {
		return accounts
	}
	if err := json.Unmarshal([]byte(*u.Accounts), &accounts); err != nil {
		log.Printf("***> unexpected error occurred when unmarshalling user accounts error: %v\n", err)
	}
	return accounts
}

func (u User) RemoveAccount(accountID string) map[string]interface{} {
	accounts := u.AccountsB()
	delete(accounts, accountID)
	return accounts
}

func (u User) AddAccount(accountID, memberID string) map[string]interface{} {
	accounts := u.AccountsB()
	accounts[accountID] = memberID
	return accounts
}

func (u User) AccountIDs() []string {
	accounts := u.AccountsB()
	accoutIds := make([]string, 0)
	for k := range accounts {
		accoutIds = append(accoutIds, k)
	}
	return accoutIds
}
