package mid

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/platform/auth"
	"gitlab.com/vjsideprojects/relay/internal/platform/redisdb"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
	"gitlab.com/vjsideprojects/relay/internal/user"
	"gitlab.com/vjsideprojects/relay/internal/visitor"
	"go.opencensus.io/trace"
)

// ErrForbidden is returned when an authenticated user does not have a
// sufficient role for an action.
var ErrForbidden = web.NewRequestError(
	errors.New("you are not authorized for doing that action"),
	http.StatusForbidden,
)

// Authenticate validates a JWT from the `Authorization` header.
func Authenticate(authenticator *auth.Authenticator) web.Middleware {
	// This is the actual middleware function to be executed.
	f := func(after web.Handler) web.Handler {

		// Wrap this handler around the next one provided.
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
			ctx, span := trace.StartSpan(ctx, "internal.mid.Authenticate")
			defer span.End()

			// Parse the authorization header. Expected header is of
			// the format `Bearer <token>`.
			parts := strings.Split(r.Header.Get("Authorization"), " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				err := errors.New("expected authorization header format: Bearer <token>")
				return web.NewRequestError(err, http.StatusUnauthorized)
			}

			claims, err := authenticator.ParseClaims(parts[1])
			if err != nil {
				return web.NewRequestError(err, http.StatusUnauthorized)
			}
			// Add claims to the context so they can be retrieved later.
			ctx = context.WithValue(ctx, auth.Key, claims)

			return after(ctx, w, r, params)
		}

		return h
	}

	return f
}

func HasSocketAccess(rp *redis.Pool) web.Middleware {

	// This is the actual middleware function to be executed.
	f := func(after web.Handler) web.Handler {

		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
			ctx, span := trace.StartSpan(ctx, "internal.mid.HasSocketAccess")
			defer span.End()

			token := params["token"]
			userID, err := redisdb.RedisGet(rp, token)
			if err != nil {
				return web.NewRequestError(err, http.StatusUnauthorized)
			}
			// Add claims to the context so they can be retrieved later.
			ctx = context.WithValue(ctx, auth.SocketKey, userID)

			return after(ctx, w, r, params)
		}

		return h
	}

	return f
}

// HasRole validates that an authenticated user has at least one role from a
// specified list. This method constructs the actual function that is used.
func HasRole(roles ...string) web.Middleware {

	// This is the actual middleware function to be executed.
	f := func(after web.Handler) web.Handler {

		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
			ctx, span := trace.StartSpan(ctx, "internal.mid.HasRole")
			defer span.End()

			claims, ok := ctx.Value(auth.Key).(auth.Claims)
			if !ok {
				err := errors.New("claims missing from context: HasRole called without/before Authenticate")
				return web.NewRequestError(err, http.StatusUnauthorized)
			}

			if !claims.HasRole(roles...) {
				return ErrForbidden
			}

			return after(ctx, w, r, params)
		}

		return h
	}

	return f
}

// HasAccountAccess validates that an authenticated user has access to the account
func HasAccountAccess(db *sqlx.DB) web.Middleware {

	// This is the actual middleware function to be executed.
	f := func(after web.Handler) web.Handler {

		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
			ctx, span := trace.StartSpan(ctx, "internal.mid.HasAccountAccess")
			defer span.End()

			claims, ok := ctx.Value(auth.Key).(auth.Claims)
			if !ok {
				return errors.New("claims missing from context: HasAccountAccess called without/before Authenticate")
			}

			accountID := params["account_id"]
			userID := claims.Subject
			usr, err := user.RetrieveUser(ctx, db, userID)
			if err != nil {
				err := errors.New("user_not_exist") // value used in the UI dont change the string message.
				return web.NewRequestError(err, http.StatusForbidden)
			}

			if hasRoleAdmin(usr.Roles) || hasRoleUser(usr.Roles) {
				if !isExist(usr.AccountIDs(), accountID) {
					err := errors.New("account_not_associated_with_this_user") // value used in the UI dont change the string message.
					return web.NewRequestError(err, http.StatusForbidden)
				}
			} else if hasRoleVisitor(usr.Roles) {
				entityID := params["entity_id"]
				itemID := params["item_id"]
				vl, err := visitor.ListByEmail(ctx, accountID, usr.Email, db)
				if err != nil {
					err := errors.New("account_not_associated_with_this_visitor") // value used in the UI dont change the string message.
					return web.NewRequestError(err, http.StatusForbidden)
				}
				hasAccess := false
				for _, vi := range vl {
					if vi.AccountID == accountID && vi.EntityID == entityID && vi.ItemID == itemID {
						hasAccess = true
						break
					}
				}
				if !hasAccess {
					err := errors.New("module_not_associated_with_this_visitor") // value used in the UI dont change the string message.
					return web.NewRequestError(err, http.StatusForbidden)
				}

				log.Println("VISITOR LOGGED IN")
				log.Println("Check for access ", entityID)
				log.Println("Check for access ", itemID)
			}

			return after(ctx, w, r, params)
		}

		return h
	}

	return f
}

func HasSlackAccess(authenticator *auth.Authenticator) web.Middleware {
	// This is the actual middleware function to be executed.
	f := func(after web.Handler) web.Handler {

		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
			ctx, span := trace.StartSpan(ctx, "internal.mid.HasSlackAccess")
			defer span.End()

			type Payload struct {
				Token     string `json:"token"`
				Type      string `json:"type"`
				Challenge string `json:"challenge"`
			}

			var sp Payload
			b, err := ioutil.ReadAll(r.Body)
			r.Body = ioutil.NopCloser(bytes.NewReader(b))
			//fmt.Println("SLACK BODY RAW : ", string(b))
			if err != nil {
				err := errors.New("brain middleware unable to read slack event body") // value used in the UI dont change the string message.
				return web.NewRequestError(err, http.StatusBadRequest)
			}
			err = json.Unmarshal(b, &sp)
			if err != nil {
				err := errors.New("brain middleware unable to unmarshal slack event body") // value used in the UI dont change the string message.
				return web.NewRequestError(err, http.StatusBadRequest)
			}

			//challenge from slack should stop here to respond.
			if sp.Challenge != "" {
				var slackChallengeRes struct {
					Challenge string `json:"challenge"`
				}
				slackChallengeRes.Challenge = sp.Challenge
				return web.Respond(ctx, w, slackChallengeRes, http.StatusOK)
			}

			// continue with slack auth checks here....

			if hasValidSlackSigningSecret(r, authenticator.SlackSignature, string(b)) != nil {
				return web.NewRequestError(err, http.StatusForbidden)
			}

			return after(ctx, w, r, params)
		}

		return h
	}

	return f
}

func hasValidSlackSigningSecret(r *http.Request, slackSignature, body string) error {
	hasedSlackSignature := r.Header.Get("X-Slack-Signature")
	slackReqTs, err := strconv.ParseInt(r.Header.Get("X-Slack-Request-Timestamp"), 10, 64)
	if err != nil {
		return err
	}
	if time.Now().UTC().Unix()-slackReqTs > 60*5 {
		//It could be a replay attack, so let's ignore it.
		return errors.New("slack ts not recent")
	}

	sigBasestring := fmt.Sprintf("%s:%d:%s", "v0", slackReqTs, body)
	mySignature := fmt.Sprintf("%s:%s", "v0=", hmac256(slackSignature, sigBasestring))

	if mySignature == hasedSlackSignature {
		return errors.New("signature mismatch")
	}

	return nil
}

func hasRoleAdmin(roles []string) bool {
	for _, r := range roles {
		if r == auth.RoleAdmin {
			return true
		}
	}
	return false
}

func hasRoleUser(roles []string) bool {
	for _, r := range roles {
		if r == auth.RoleUser {
			return true
		}
	}
	return false
}

func hasRoleVisitor(roles []string) bool {
	for _, r := range roles {
		if r == auth.RoleUser {
			return true
		}
	}
	return false
}

func isExist(accountIDs []string, accountIDInReqParam string) bool {
	for _, accountID := range accountIDs {
		if accountIDInReqParam == accountID {
			return true
		}
	}
	return false
}

func hmac256(secret, data string) string {
	h := hmac.New(sha256.New, []byte(secret))
	// Write Data to it
	h.Write([]byte(data))
	// Get result and encode as hexadecimal string
	return hex.EncodeToString(h.Sum(nil))

}
