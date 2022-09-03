package team

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"go.opencensus.io/trace"
)

var (
	// ErrNotFound is used when a specific entity is requested but does not exist.
	ErrNotFound = errors.New("Team not found")

	// ErrInvalidID occurs when an ID is not in a valid form.
	ErrInvalidID = errors.New("ID is not in its proper form")
)

func Map(ctx context.Context, accountID string, db *sqlx.DB) (map[string]Team, error) {
	teams, err := List(ctx, accountID, db)
	if err != nil {
		return nil, err
	}
	teamMap := make(map[string]Team, 0)
	for _, t := range teams {
		teamMap[t.ID] = t
	}
	return teamMap, nil
}

// List retrieves a list of existing teams for the account associated from the database.
func List(ctx context.Context, accountID string, db *sqlx.DB) ([]Team, error) {
	ctx, span := trace.StartSpan(ctx, "internal.team.List")
	defer span.End()

	teams := []Team{}
	const q = `SELECT * FROM teams where account_id = $1 LIMIT 50`

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
		ID:          uuid.New().String(),
		AccountID:   n.AccountID,
		Name:        n.Name,
		Description: &n.Description,
		CreatedAt:   now.UTC(),
		UpdatedAt:   now.UTC().Unix(),
	}

	const q = `INSERT INTO teams
		(team_id, account_id, name, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := db.ExecContext(
		ctx, q,
		t.ID, t.AccountID, t.Name, t.Description,
		t.CreatedAt, t.UpdatedAt,
	)

	if err != nil {
		return Team{}, errors.Wrap(err, "inserting team")
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

func CustomModules() []Module {
	modules := make([]Module, 0)
	for k, v := range modulesMap {
		module := Module{
			Key:  k,
			Name: v,
		}
		modules = append(modules, module)
	}
	return modules
}

func CustomTemplates() []Template {
	templates := make([]Template, 0)
	for k, v := range templatesMap {
		template := Template{
			Key:         k,
			Name:        v,
			Description: templatesDescMap[k],
		}
		templates = append(templates, template)
	}
	return templates
}

func Names(teams []Team) []string {
	names := make([]string, 0)
	for _, t := range teams {
		if t.AccountID != t.ID { //skip base
			names = append(names, t.Name)
		}
	}
	return names
}
