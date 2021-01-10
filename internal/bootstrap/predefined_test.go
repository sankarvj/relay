package bootstrap_test

import (
	"context"
	"testing"

	"github.com/dgrijalva/jwt-go"
	"gitlab.com/vjsideprojects/relay/internal/bootstrap"
	"gitlab.com/vjsideprojects/relay/internal/platform/auth"
	"gitlab.com/vjsideprojects/relay/internal/schema"
	"gitlab.com/vjsideprojects/relay/internal/tests"
)

func TestSaveTokenOnOwnerEntity(t *testing.T) {
	db, teardown := tests.NewUnit(t)
	defer teardown()
	tests.SeedData(t, db)
	tests.SeedEntity(t, db)
	tests.SeedRelationShips(t, db)

	t.Log(" Save Token To The User")
	{
		t.Log("\tWhen adding token to user entity")
		{
			ctx := context.WithValue(tests.Context(), 1, &auth.Claims{
				Roles: []string{"ADMIN", "USER"},
				StandardClaims: jwt.StandardClaims{
					Subject:   "5cf37266-3473-4006-984f-9325122678b7",
					IssuedAt:  1610259806,
					ExpiresAt: 1610346206,
					NotBefore: 0,
					Audience:  "",
					Issuer:    "",
				},
			})

			err := bootstrap.SaveToken(ctx, schema.SeedAccountID, "token", db)
			if err != nil {
				t.Fatalf("\tCould not able to save token on current user - %s", err)
			}
		}
	}
}
