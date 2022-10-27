package dashboard

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"go.opencensus.io/trace"
)

var (
	// ErrDashboardNotFound is used when a specific dashboard is requested but does not exist.
	ErrDashboardNotFound = errors.New("Dashboard not found")
)

func List(ctx context.Context, accountID, teamID, entityID string, db *sqlx.DB) ([]Dashboard, error) {
	ctx, span := trace.StartSpan(ctx, "internal.dashboard.List")
	defer span.End()

	dashboards := []Dashboard{}
	const q = `SELECT * FROM dashboards where account_id = $1 AND team_id = $2 AND entity_id = $3 LIMIT 50`
	if err := db.SelectContext(ctx, &dashboards, q, accountID, teamID, entityID); err != nil {
		return dashboards, errors.Wrap(err, "selecting dashboards for an account")
	}

	return dashboards, nil
}

func Create(ctx context.Context, db *sqlx.DB, nd NewDashboard, now time.Time) error {
	ctx, span := trace.StartSpan(ctx, "internal.dashboard.Create")
	defer span.End()

	metaBytes, err := json.Marshal(nd.Meta)
	if err != nil {
		return errors.Wrap(err, "encode meta to bytes in dashboard")
	}

	d := Dashboard{
		ID:        nd.ID,
		AccountID: nd.AccountID,
		TeamID:    nd.TeamID,
		UserID:    nd.UserID,
		EntityID:  nd.EntityID,
		Name:      nd.Name,
		Type:      nd.Type,
		Metab:     string(metaBytes),
		CreatedAt: now.UTC(),
	}

	const q = `INSERT INTO dashboards
		(dashboard_id, account_id, team_id, user_id, entity_id, name, type, metab, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err = db.ExecContext(
		ctx, q,
		d.ID, d.AccountID, d.TeamID, d.UserID, d.EntityID, d.Name, d.Type, d.Metab,
		d.CreatedAt,
	)
	if err != nil {
		return errors.Wrap(err, "dashboard created")
	}

	return nil
}

func Delete(ctx context.Context, db *sqlx.DB, accountID, teamID, dashboardID string) error {
	ctx, span := trace.StartSpan(ctx, "internal.dashboard.Delete")
	defer span.End()

	const q = `DELETE FROM dashboards WHERE account_id = $1 AND team_id = $2 AND dashboard_id = $3`

	if _, err := db.ExecContext(ctx, q, accountID, teamID, dashboardID); err != nil {
		return errors.Wrapf(err, "dashboard delete")
	}

	return nil
}

func Retrieve(ctx context.Context, accountID, teamID, dashboardID string, db *sqlx.DB) (*Dashboard, error) {
	ctx, span := trace.StartSpan(ctx, "internal.dashboard.Retrieve")
	defer span.End()

	var d Dashboard
	const q = `SELECT * FROM dashboards WHERE account_id = $1 AND team_id = $2 AND dashboard_id = $3`
	if err := db.GetContext(ctx, &d, q, accountID, teamID, dashboardID); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrDashboardNotFound
		}

		return nil, err
	}

	return &d, nil
}

func RetrieveByEntity(ctx context.Context, accountID, teamID, entityID string, db *sqlx.DB) (*Dashboard, error) {
	ctx, span := trace.StartSpan(ctx, "internal.dashboard.RetrieveByBaseEntity")
	defer span.End()

	var d Dashboard
	const q = `SELECT * FROM dashboards WHERE account_id = $1 AND team_id = $2 AND entity_id = $3 LIMIT 1`
	if err := db.GetContext(ctx, &d, q, accountID, teamID, entityID); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrDashboardNotFound
		}

		return nil, err
	}

	return &d, nil
}

func (d Dashboard) Meta() map[string]string {
	meta := make(map[string]string, 0)
	if d.Metab == "" {
		return meta
	}
	if err := json.Unmarshal([]byte(d.Metab), &meta); err != nil {
		log.Printf("***> unexpected error occurred when unmarshalling meta for dashboard: %v error: %v\n", d.ID, err)
	}
	return meta
}

func BuildNewDashboard(accountID, teamID, userID, entityID, name string) *NewDashboard {
	return &NewDashboard{
		ID:        uuid.New().String(),
		AccountID: accountID,
		TeamID:    teamID,
		UserID:    userID,
		EntityID:  entityID,
		Name:      name,
		Type:      string(TypeDefault),
		Meta:      map[string]string{},
	}
}

func (nd *NewDashboard) Add(ctx context.Context, db *sqlx.DB) (string, error) {
	err := Create(ctx, db, *nd, time.Now())
	if err != nil {
		return "", err
	}
	return nd.ID, nil
}
