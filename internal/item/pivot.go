package item

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
	rg "github.com/redislabs/redisgraph-go"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"go.opencensus.io/trace"
)

func graph(graphName string, conn redis.Conn) rg.Graph {
	return rg.GraphNew(graphName, conn)
}

//AddItemNode adds the graph entry for each item(node)
func AddItemNode(ctx context.Context, rPool *redis.Pool, accountID, entityID string, fields []entity.Field, now time.Time) error {
	ctx, span := trace.StartSpan(ctx, "internal.item.AddItemNode")
	defer span.End()

	conn := rPool.Get()
	defer conn.Close()

	graph := graph(accountID, conn)

	//---------------------------
	japan := rg.Node{
		Alias: "l",
		Label: "country",
		Properties: map[string]interface{}{
			"name": "Japan",
		},
	}
	graph.AddNode(&japan)
	properties := map[string]interface{}{}
	for _, field := range fields {
		if field.List {
			list := field.Value.([]interface{})
			for _, v := range list {
				lNode := rg.Node{
					Alias: "l",
					Label: "lists",
					Properties: map[string]interface{}{
						"value": v,
					},
				}
				graph.AddNode(&lNode)
			}
		} else {
			properties[quote(field.Key)] = field.Value
		}
	}
	//---------------------------
	iNode := rg.Node{
		Alias:      "i",
		Label:      quote(entityID),
		Properties: properties,
	}
	graph.AddNode(&iNode)

	for _, n := range graph.Nodes {
		if n.Alias == "l" {
			edge := rg.Edge{
				Source:      &iNode,
				Relation:    "visited",
				Destination: n,
			}
			graph.AddEdge(&edge)
		}
	}

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

//Upsert .....
// func Upsert(rPool *redis.Pool, accountID, alias, label string, properties map[string]interface{}) error {
// 	conn := rPool.Get()
// 	defer conn.Close()
// 	graph := graph(accountID, conn)
// 	iNode, err := CreateIfNotExists(rPool, accountID, alias, label, properties)
// 	if err != nil {
// 		return err
// 	}
// 	log.Println("iNode", iNode)
// 	log.Println("graph", graph)
// 	return nil
// }

//Upsert hi
func Upsert(rPool *redis.Pool, accountID, label, id string, newFields []entity.Field) (*rg.Node, error) {
	conn := rPool.Get()
	defer conn.Close()
	graph := graph(accountID, conn)

	iNode, err := wrapFindNode(rPool, accountID, label, id)
	if err != nil {
		return nil, err
	}

	err = setProperties(&graph, iNode, newFields)

	Temp(rPool, accountID, label, quote("4d247443-b257-4b06-ba99-493cf9d83ce7"), id)

	return iNode, errors.Wrap(err, "err on commit")
}

func wrapFindNode(rPool *redis.Pool, accountID, label, id string) (*rg.Node, error) {
	iNode, err := getNode(rPool, accountID, label, id)
	if err != nil {
		return nil, err
	}
	if iNode == nil {
		return &rg.Node{
			Label: quote(label),
			Properties: map[string]interface{}{
				"id": id,
			},
		}, nil
	}
	return iNode, nil
}

func getNode(rPool *redis.Pool, accountID, label, id string) (*rg.Node, error) {
	log.Printf("get node for %v", id)
	conn := rPool.Get()
	defer conn.Close()
	graph := graph(accountID, conn)

	var iNode *rg.Node
	query := fmt.Sprintf(`MATCH (i:%s) where i.id = "%s" return i`, quote(label), id)
	result, err := graph.Query(query)
	if err != nil {
		return iNode, errors.Wrap(err, "selection nodes")
	}
	result.PrettyPrint()
	if result.Next() {
		records := result.Record().Values()
		keys := result.Record().Keys()
		if result.Next() || len(records) == 0 {
			return iNode, errors.New("problem getting the node for id: " + id)
		}
		iNode = records[0].(*rg.Node)
		log.Printf("keys --> %+v", keys)
		log.Printf("n.ID --> %+v", iNode)
		log.Printf("result --> %+v", result)
		iNode.Label = label
	}
	return iNode, nil
}

func Temp(rPool *redis.Pool, accountID, label, label1, id string) (*rg.Node, error) {
	log.Printf("calling temp for %v", id)
	iNode := rg.Node{
		Label: label,
	}

	conn := rPool.Get()
	defer conn.Close()
	graph := graph(accountID, conn)

	query := fmt.Sprintf(`MATCH (i:%s)-[c:contains]->(l:%s{element:"yellow"}) where i.id = "%s" return i,l`, quote(label), quote(label1), id)
	result, err := graph.Query(query)
	if err != nil {
		return nil, errors.Wrap(err, "selection nodes ...")
	}
	log.Printf("calling temp result %v", result)
	result.PrettyPrint()
	if result.Next() {
		records := result.Record().Values()
		if result.Next() || len(records) == 0 {
			return nil, errors.New("problem getting the node for id: " + id)
		}
		n := records[0].(*rg.Node)
		iNode.Properties = n.Properties
	}
	return &iNode, nil
}

func Temp2(rPool *redis.Pool, accountID, label, label1, id string) (*rg.Node, error) {
	log.Printf("calling temp1 for %v", id)
	iNode := rg.Node{
		Label: quote(label),
	}

	conn := rPool.Get()
	defer conn.Close()
	graph := graph(accountID, conn)

	query := fmt.Sprintf(`MATCH (i:%s)-[c:contains]->(l:%s) where i.id = "%s" return i,l`, quote(label), quote(label1), id)
	result, err := graph.Query(query)
	if err != nil {
		return nil, errors.Wrap(err, "selection nodes ...")
	}
	result.PrettyPrint()
	if result.Next() {
		records := result.Record().Values()
		if result.Next() || len(records) == 0 {
			return nil, errors.New("problem getting the node for id: " + id)
		}
		n := records[0].(*rg.Node)
		iNode.Properties = n.Properties
	}
	return &iNode, nil
}

func setProperties(graph *rg.Graph, iNode *rg.Node, newFields []entity.Field) error {
	isModified := false
	graph.AddNode(iNode)

	for _, field := range newFields {
		if val, ok := iNode.Properties[field.Key]; (ok && field.Value == val) || field.Key == "id" {
			continue
		}
		isModified = true
		if !createEdges(graph, iNode, field) { // don't add properties if edges exists
			iNode.Properties[quote(field.Key)] = field.Value
		}
	}
	if isModified {
		_, err := Commit(graph)
		if err != nil {
			return err
		}
	}
	return nil
}

//TODO delete old edges/nodes of the list/reference
func createEdges(graph *rg.Graph, sNode *rg.Node, field entity.Field) bool {
	if field.List {
		list := field.Value.([]string)

		for _, element := range list {
			log.Printf("element %v", element)
			dNode := rg.Node{
				Label: quote(field.Key),
				Properties: map[string]interface{}{
					"element": element,
				},
			}
			graph.AddNode(&dNode) // Merge??? / get lists / delete and recreate :(
			edge := rg.Edge{
				Source:      sNode,
				Relation:    "contains",
				Destination: &dNode,
			}
			log.Printf("edge added %+v", edge)
			graph.AddEdge(&edge)
		}
		return true
	} else if field.DataType == entity.TypeReference {
		//call getNode and attach
		return true
	}
	return false
}

//Commit ...
func Commit(g *rg.Graph) (*rg.QueryResult, error) {
	items := make([]string, 0, len(g.Nodes)+len(g.Edges))
	for _, n := range g.Nodes {
		items = append(items, n.Encode())
	}
	for _, e := range g.Edges {
		items = append(items, e.Encode())
	}
	log.Println("items -----> ", items)
	q := "CREATE " + strings.Join(items, ",")
	log.Println("qqqqqqqq -----> ", q)
	return g.Query(q)
}

func Update(rPool *redis.Pool, accountID, label, id string, properties map[string]interface{}) (*rg.QueryResult, error) {
	conn := rPool.Get()
	defer conn.Close()
	graph := graph(accountID, conn)

	q := updateProperties(label, id, properties)

	log.Println("Update qqqqqqqq -----> ", q)
	return graph.Query(q)
}

func UpdateRelation(rPool *redis.Pool, accountID, label, id string) (*rg.QueryResult, error) {
	conn := rPool.Get()
	defer conn.Close()
	graph := graph(accountID, conn)

	srcNode := newNode(label, rg.RandomString(10), id)

	dNode := &rg.Node{
		Alias: rg.RandomString(10),
		Label: "lists",
		Properties: map[string]interface{}{
			"element": "honda",
		},
	}

	q := updateListPath("contains", srcNode, dNode)

	log.Println("UpdateRelation qqqqqqqq -----> ", q)
	return graph.Query(q)
}

//"MATCH (n { id: '12345' }) SET n.age = 33, n.name = 'Bob'"
func updateProperties(label, id string, properties map[string]interface{}) string {
	alias := rg.RandomString(10)
	n := newNode(label, alias, id)
	s := matchNode(n)
	if len(properties) > 0 {
		p := make([]string, 0, len(properties))
		for k, v := range properties {
			if k == "id" {
				continue
			}
			p = append(p, fmt.Sprintf("%s.%s = %v", alias, k, rg.ToString(v)))
		}

		s = append(s, "SET")
		s = append(s, " ")
		s = append(s, strings.Join(p, ","))
	}

	return strings.Join(s, "")
}

//"MERGE (charlie { name: 'Charlie Sheen' }) MERGE (wallStreet:Movie { name: 'Wall Street' }) MERGE (charlie)-[r:ACTED_IN]->(wallStreet)"
func updateListPath(relation string, srcNode, destNode *rg.Node) string {
	edge := rg.EdgeNew(relation, srcNode, destNode, nil)
	ms := matchNode(srcNode)
	md := mergeNode(destNode)
	me := mergeEdge(edge)
	return strings.Join(ms, "") + strings.Join(md, "") + strings.Join(me, "")

}

func newNode(label, alias, id string) *rg.Node {
	n := rg.NodeNew(quote(label), alias, map[string]interface{}{
		"id": id,
	})
	return n
}

func matchNode(n *rg.Node) []string {
	s := []string{"MATCH"}
	s = append(s, " ")
	s = append(s, n.Encode())
	s = append(s, " ")
	return s
}

func mergeNode(n *rg.Node) []string {
	s := []string{"MERGE"}
	s = append(s, " ")
	s = append(s, n.Encode())
	s = append(s, " ")
	return s
}

func mergeEdge(e *rg.Edge) []string {
	s := []string{"MERGE"}
	s = append(s, " ")
	s = append(s, e.Encode())
	s = append(s, " ")
	return s
}

func quote(label string) string {
	return fmt.Sprintf("`%s`", label)
}
