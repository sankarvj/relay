package team

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"go.opencensus.io/trace"
)

var (
	// ErrNotFound is used when a specific entity is requested but does not exist.
	ErrNotFound = errors.New("Team not found")

	// ErrInvalidID occurs when an ID is not in a valid form.
	ErrInvalidID = errors.New("ID is not in its proper form")
)

// List retrieves a list of existing teams for the account associated from the database.
func List(ctx context.Context, accountID string, db *sqlx.DB) ([]Team, error) {
	ctx, span := trace.StartSpan(ctx, "internal.team.List")
	defer span.End()

	teams := []Team{}
	const q = `SELECT * FROM teams where account_id = $1`

	if err := db.SelectContext(ctx, &teams, q, accountID); err != nil {
		return nil, errors.Wrap(err, "selecting teams")
	}
	return teams, nil
}

// Create inserts a new user into the database.
func Create(ctx context.Context, db *sqlx.DB, n NewTeam, now time.Time) (Team, error) {
	ctx, span := trace.StartSpan(ctx, "internal.team.Create")
	defer span.End()

	t := Team{
		ID:        uuid.New().String(),
		AccountID: n.AccountID,
		Name:      n.Name,
		CreatedAt: now.UTC(),
		UpdatedAt: now.UTC().Unix(),
	}

	const q = `INSERT INTO teams
		(team_id, account_id, name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)`
	_, err := db.ExecContext(
		ctx, q,
		t.ID, t.AccountID, t.Name,
		t.CreatedAt, t.UpdatedAt,
	)

	if err != nil {
		return Team{}, errors.Wrap(err, "inserting team")
	}

	ne := entity.NewEntity{
		Name:      "Members",
		AccountID: t.AccountID,
		TeamID:    t.ID,
		Category:  entity.CategoryUserSeries,
		Fields:    makeMemberSeriesFields(),
	}
	_, err = entity.Create(ctx, db, ne, now)
	if err != nil {
		return Team{}, errors.Wrap(err, "inserting members enetity of each team")
	}

	return t, nil
}

// Retrieve gets the specified entity from the database.
func Retrieve(ctx context.Context, teamID int64, db *sqlx.DB) (Team, error) {
	ctx, span := trace.StartSpan(ctx, "internal.team.Retrieve")
	defer span.End()

	var t Team
	const q = `SELECT * FROM teams WHERE team_id = $1`
	if err := db.GetContext(ctx, &t, q, teamID); err != nil {
		if err == sql.ErrNoRows {
			return Team{}, ErrNotFound
		}

		return Team{}, errors.Wrapf(err, "selecting team %q", teamID)
	}

	return t, nil
}

func makeMemberSeriesFields() []entity.Field {
	fields := make([]entity.Field, 0)
	fields = append(fields, makeNewField("Email", uuid.New().String(), "e1", false, entity.TypeString))
	fields = append(fields, makeNewField("Name", uuid.New().String(), "", false, entity.TypeString))
	fields = append(fields, makeNewField("Email", uuid.New().String(), "", false, entity.TypeString))
	return fields
}

func makeNewField(name, key, value string, hidden bool, dataType entity.DType) entity.Field {
	field := entity.Field{
		Name:     name,
		Key:      key,
		DataType: dataType,
		Value:    value,
		Unique:   true,
		Hidden:   hidden,
	}
	return field
}
