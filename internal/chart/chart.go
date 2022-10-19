package chart

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"go.opencensus.io/trace"
)

var (
	// ErrChartNotFound is used when a specific chart is requested but does not exist.
	ErrChartNotFound = errors.New("Chart not found")
)

func List(ctx context.Context, accountID, baseEntityID, entityID string, db *sqlx.DB) ([]Chart, error) {
	ctx, span := trace.StartSpan(ctx, "internal.chart.List")
	defer span.End()

	if baseEntityID == "" || baseEntityID == "undefined" {
		baseEntityID = entity.NoEntityID
	}

	if entityID == "" || entityID == "undefined" {
		entityID = entity.NoEntityID
	}

	log.Println("baseEntityID ", baseEntityID)
	log.Println("entityID ", entityID)

	charts := []Chart{}
	if entityID == entity.NoEntityID {
		const q = `SELECT * FROM charts where account_id = $1 AND base_entity_id = $2 LIMIT 50`
		if err := db.SelectContext(ctx, &charts, q, accountID, baseEntityID); err != nil {
			return charts, errors.Wrap(err, "selecting charts for an account")
		}
	} else {
		const q = `SELECT * FROM charts where account_id = $1 AND entity_id = $2 AND base_entity_id = $3 LIMIT 50`
		if err := db.SelectContext(ctx, &charts, q, accountID, entityID, baseEntityID); err != nil {
			return charts, errors.Wrap(err, "selecting charts for an account with group")
		}
	}

	return charts, nil
}

func Create(ctx context.Context, db *sqlx.DB, nc NewChart, now time.Time) error {
	ctx, span := trace.StartSpan(ctx, "internal.chart.Create")
	defer span.End()

	metaBytes, err := json.Marshal(nc.Meta)
	if err != nil {
		return errors.Wrap(err, "encode meta to bytes in chart")
	}

	t := Chart{
		ID:           uuid.New().String(),
		AccountID:    nc.AccountID,
		EntityID:     nc.EntityID,
		BaseEntityID: nc.BaseEntityID,
		UserID:       nc.UserID,
		Name:         nc.Name,
		Type:         nc.Type,
		Duration:     nc.Duration,
		State:        nc.State,
		Position:     nc.Position,
		Metab:        string(metaBytes),
		CreatedAt:    now.UTC(),
	}

	const q = `INSERT INTO charts
		(chart_id, account_id, entity_id, base_entity_id, user_id, name, type, duration, state, position, metab, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`
	_, err = db.ExecContext(
		ctx, q,
		t.ID, t.AccountID, t.EntityID, t.BaseEntityID, t.UserID, t.Name, t.Type, t.Duration, t.State, t.Position, t.Metab,
		t.CreatedAt,
	)
	if err != nil {
		return errors.Wrap(err, "chart created")
	}

	return nil
}

func Delete(ctx context.Context, db *sqlx.DB, chartID string) error {
	ctx, span := trace.StartSpan(ctx, "internal.chart.Delete")
	defer span.End()

	const q = `DELETE FROM charts WHERE chart_id = $1`

	if _, err := db.ExecContext(ctx, q, chartID); err != nil {
		return errors.Wrapf(err, "chart delete")
	}

	return nil
}

func Retrieve(ctx context.Context, accountID, chartID string, db *sqlx.DB) (*Chart, error) {
	ctx, span := trace.StartSpan(ctx, "internal.chart.Retrieve")
	defer span.End()

	var c Chart
	const q = `SELECT * FROM charts WHERE account_id =$1 AND chart_id = $2`
	if err := db.GetContext(ctx, &c, q, accountID, chartID); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrChartNotFound
		}

		return nil, err
	}

	return &c, nil
}

func (c Chart) Meta() map[string]string {
	meta := make(map[string]string, 0)
	if c.Metab == "" {
		return meta
	}
	if err := json.Unmarshal([]byte(c.Metab), &meta); err != nil {
		log.Printf("***> unexpected error occurred when unmarshalling meta for chart: %v error: %v\n", c.ID, err)
	}
	return meta
}

func (c Chart) GetGroupByLogic() string {
	if val, ok := c.Meta()[MetaGroupByLogic]; ok {
		return val
	}
	return string(GroupLogicNone)
}

func (c Chart) GetCalc() string {
	if val, ok := c.Meta()[MetaCalcKey]; ok {
		return val
	}
	return string(CalcCount)
}

func (c Chart) GetSource() string {
	if val, ok := c.Meta()[MetaSourceKey]; ok {
		return val
	}
	return ""
}

func (c Chart) GetField() string {
	if val, ok := c.Meta()[MetaFieldKey]; ok {
		return val
	}
	return ""
}

func (c Chart) GetDType() string {
	if val, ok := c.Meta()[MetaDataType]; ok {
		return val
	}
	return string(DTypeDefault)
}

func (c Chart) GetExp() string {
	if val, ok := c.Meta()[MetaExp]; ok {
		return val
	}
	return ""
}

func (c Chart) GetDate() string {
	if val, ok := c.Meta()[MetaDateField]; ok {
		return val
	}
	return ""
}

func BuildNewChart(accountID, userID, entityID, name, fieldName string, chartType Type) *NewChart {
	NoEntityID := "00000000-0000-0000-0000-000000000000"
	return &NewChart{
		AccountID:    accountID,
		EntityID:     entityID,
		BaseEntityID: entity.NoEntityID, // this is useful to categorize charts based on entity.
		UserID:       userID,
		Name:         name,
		Type:         string(chartType),
		Duration:     string(LastWeek),
		Meta: map[string]string{
			MetaSourceKey:    NoEntityID,
			MetaFieldKey:     fieldName,
			MetaDataType:     string(DTypeDefault),
			MetaCalcKey:      string(CalcCount),
			MetaGroupByLogic: string(GroupLogicNone),
		},
	}
}

func (ch *NewChart) AddSource(source string) *NewChart {
	ch.Meta[MetaSourceKey] = source
	return ch
}
func (ch *NewChart) AddDateField(fieldDate string) *NewChart {
	ch.Meta[MetaDateField] = fieldDate
	return ch
}
func (ch *NewChart) AddExp(exp string) *NewChart {
	ch.Meta[MetaExp] = exp
	return ch
}

func (ch *NewChart) SetGrpLogicID() *NewChart {
	ch.Meta[MetaGroupByLogic] = string(GroupLogicID)
	return ch
}
func (ch *NewChart) SetGrpLogicField() *NewChart {
	ch.Meta[MetaGroupByLogic] = string(GroupLogicField)
	return ch
}
func (ch *NewChart) SetGrpLogicParent() *NewChart {
	ch.Meta[MetaGroupByLogic] = string(GroupLogicParent)
	return ch
}

func (ch *NewChart) SetAsTimeseries() *NewChart {
	ch.Meta[MetaDataType] = string(DTypeTimeseries)
	return ch
}

func (ch *NewChart) SetAsCustom() *NewChart {
	ch.Meta[MetaDataType] = string(DTypeCustom)
	return ch
}

func (ch *NewChart) SetDurationAllTime() *NewChart {
	ch.Duration = string(AllTime)
	return ch
}
func (ch *NewChart) SetDurationLast24hrs() *NewChart {
	ch.Duration = string(Last24Hrs)
	return ch
}

func (ch *NewChart) SetCalcRate() *NewChart {
	ch.Meta[MetaCalcKey] = string(CalcRate)
	return ch
}
func (ch *NewChart) SetCalcSum() *NewChart {
	ch.Meta[MetaCalcKey] = string(CalcSum)
	return ch
}

func (ch *NewChart) SetBaseEntityID(baseEntityID string) *NewChart {
	ch.BaseEntityID = baseEntityID // useful for filtering.
	return ch
}

func (ch *NewChart) Add(ctx context.Context, db *sqlx.DB) error {
	err := Create(ctx, db, *ch, time.Now())
	if err != nil {
		return err
	}
	return nil
}
