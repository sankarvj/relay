package auth

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/platform/web"
	"gitlab.com/vjsideprojects/relay/internal/relationship"
	"gitlab.com/vjsideprojects/relay/internal/visitor"
)

// These are the expected values for Claims.Roles.
const (
	RoleAdmin   = "ADMIN"
	RoleMember  = "MEMBER" // not yet implemented
	RoleUser    = "USER"
	RoleVisitor = "VISITOR"
)

// ctxKey represents the type of value for the context key.
type ctxKey int

// Key is used to store/retrieve a Claims value from a context.Context.
const Key ctxKey = 1
const SocketKey ctxKey = 100
const RoleKey ctxKey = 200

// Claims represents the authorization claims transmitted via a JWT.
type Claims struct {
	Roles []string `json:"roles"`
	jwt.StandardClaims
}

// NewClaims constructs a Claims value for the identified user. The Claims
// expire within a specified duration of the provided time. Additional fields
// of the Claims can be set after calling NewClaims is desired.
func NewClaims(subject string, roles []string, now time.Time, expires time.Duration) Claims {
	c := Claims{
		Roles: roles,
		StandardClaims: jwt.StandardClaims{
			Subject:   subject,
			IssuedAt:  now.Unix(),
			ExpiresAt: now.Add(expires).Unix(),
		},
	}

	return c
}

// Valid is called during the parsing of a token.
func (c Claims) Valid() error {
	for _, r := range c.Roles {
		switch r {
		case RoleAdmin, RoleUser, RoleVisitor: // Role is valid.
		default:
			return fmt.Errorf("invalid role %q", r)
		}
	}
	if err := c.StandardClaims.Valid(); err != nil {
		return errors.Wrap(err, "validating standard claims")
	}
	return nil
}

// HasRole returns true if the claims has at least one of the provided roles.
func (c Claims) HasRole(allowed ...string) bool {
	for _, has := range c.Roles {
		for _, alwed := range allowed {
			if has == alwed {
				return true
			}
		}
	}
	return false
}

func CheckVisitorEntityAccess(ctx context.Context, email, accountID, teamID, baseEntityID, baseItemID, entityID, itemID string, db *sqlx.DB) error {

	// baseEntityID := r.URL.Query().Get("be")
	// entityID := params["entity_id"]
	// itemID := params["item_id"]
	//usr.Email
	vl, err := visitor.ListByEmail(ctx, accountID, email, db)
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
	//re-evaluate with the base entityID - allow if the
	if !hasAccess {
		for _, vi := range vl {
			if vi.AccountID == accountID && vi.EntityID == baseEntityID {
				bonds, err := relationship.List(ctx, db, accountID, teamID, baseEntityID, false)
				if err != nil {
					return err
				}
				for _, bond := range bonds {
					log.Printf("bond ------ > %+v", bond)
					if bond.EntityID == entityID && bond.IsPublic {
						hasAccess = true
						break
					}
				}
			}
		}
	}

	if !hasAccess {
		err := errors.New("module_not_associated_with_this_visitor") // value used in the UI dont change the string message.
		return web.NewRequestError(err, http.StatusForbidden)
	}

	log.Println("VISITOR LOGGED IN")
	return nil
}

func isRoleAdmin(roles []string) bool {
	for _, r := range roles {
		if r == RoleAdmin {
			return true
		}
	}
	return false
}

func isRoleMember(roles []string) bool {
	for _, r := range roles {
		if r == RoleMember {
			return true
		}
	}
	return false
}

func isRoleUser(roles []string) bool {
	for _, r := range roles {
		if r == RoleUser {
			return true
		}
	}
	return false
}

func IsRoleVisitor(roles []string) bool {
	for _, r := range roles {
		if r == RoleVisitor {
			return true
		}
	}
	return false
}

func God(ctx context.Context) bool {
	role, ok := ctx.Value(RoleKey).(string)
	if ok {
		if role == RoleAdmin || role == RoleMember {
			return true
		}
	}
	return false
}
