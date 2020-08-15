package item

import (
	"context"
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
	rg "github.com/redislabs/redisgraph-go"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"go.opencensus.io/trace"
)

//AddItemNode adds the graph entry for each item(node)
func AddItemNode(ctx context.Context, rPool *redis.Pool, accountID, entityID string, fields []entity.Field, now time.Time) error {
	ctx, span := trace.StartSpan(ctx, "internal.item.AddItemNode")
	defer span.End()

	conn := rPool.Get()
	defer conn.Close()

	graph := graph(accountID, conn)

	iNode := rg.Node{
		Alias:      "i",
		Label:      entityID,
		Properties: fields,
	}
	graph.AddNode(&iNode)

	japan := rg.Node{
		Label: "country",
		Properties: map[string]interface{}{
			"name": "Japan",
		},
	}
	graph.AddNode(&japan)

	edge := rg.Edge{
		Source:      &iNode,
		Relation:    "visited",
		Destination: &japan,
	}
	graph.AddEdge(&edge)

	_, err := graph.Commit()
	if err != nil {
		return errors.Wrap(err, "graph commit")
	}

	query := fmt.Sprintf(`MATCH (i:%s)-[v:visited]->(c:country)
	RETURN i.id, i.name, i.age, c.name`, iNode.Label)

	// result is a QueryResult struct containing the query's generated records and statistics.
	result, err := graph.Query(query)
	if err != nil {
		return err
	}

	// Pretty-print the full result set as a table.
	result.PrettyPrint()

	//------------------------------------------------------- testing
	query = fmt.Sprintf(`MATCH (i:%s) where i.id = "%s" return i`, iNode.Label, "12345")
	// result is a QueryResult struct containing the query's generated records and statistics.
	result, err = graph.Query(query)
	if err != nil {
		return err
	}

	// Pretty-print the full result set as a table.
	result.PrettyPrint()
	//------------------------------------------------------- testing

	return nil
}

func graph(graphName string, conn redis.Conn) rg.Graph {
	return rg.GraphNew(graphName, conn)
}
