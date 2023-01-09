package dbservice

import (
	"context"

	"github.com/jmoiron/sqlx"
	"gitlab.com/vjsideprojects/relay/internal/item"
	"gitlab.com/vjsideprojects/relay/internal/platform/database"
	"gitlab.com/vjsideprojects/relay/internal/platform/graphdb"
)

type DBIdenfier string

const (
	Bee    string = "psql"
	Spider string = "redis_graph"
)

type Counters struct {
	ID      string
	GroupID string
	Count   interface{}
}

type DBService interface {
	Result(ctx context.Context, accountID, entityID, sortby, direction string, page int, docount, useReturn bool, conditions []graphdb.Field) ([]item.Item, map[string]int, error)
	Count(ctx context.Context, accountID, entityID, groupByKey, groupById, groupLogic string, conditions []graphdb.Field) ([]Counters, error)
	Sum(ctx context.Context, accountID, entityID, groupById string, conditions []graphdb.Field) ([]Counters, error)
	Search1(ctx context.Context, accountID, entityID string, conditionFields []graphdb.Field) []interface{}
	Search2(ctx context.Context, accountID, entityID string, conditionFields []graphdb.Field) ([]item.Item, error)
	Search3(ctx context.Context, accountID, entityID string, conditionFields []graphdb.Field) []interface{}
}

func CreateBeeService(pdb *sqlx.DB, sdb *database.SecDB) DBService {
	return BeeService{
		pdb: pdb,
		sdb: sdb,
	}
}

func CreateSpiderService(pdb *sqlx.DB, sdb *database.SecDB) DBService {
	return SpiderService{
		pdb: pdb,
		sdb: sdb,
	}
}

func NewDBservice(dbIdentifier string, pdb *sqlx.DB, sdb *database.SecDB) DBService {
	switch dbIdentifier {
	case Bee:
		return CreateBeeService(pdb, sdb)
	case Spider:
		return CreateSpiderService(pdb, sdb)
	default:
		return CreateBeeService(pdb, sdb)
	}
}

func FetchIds(items []item.Item) []string {
	ids := make([]string, len(items))
	for index, it := range items {
		ids[index] = it.ID
	}
	return ids
}
