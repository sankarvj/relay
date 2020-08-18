package item

import (
	"fmt"
	"log"
	"strings"

	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
	rg "github.com/redislabs/redisgraph-go"
	"gitlab.com/vjsideprojects/relay/internal/entity"
)

func graph(graphName string, conn redis.Conn) rg.Graph {
	return rg.GraphNew(graphName, conn)
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
// func Upsert(rPool *redis.Pool, accountID, label, id string, newFields []entity.Field) (*rg.Node, error) {
// 	conn := rPool.Get()
// 	defer conn.Close()
// 	graph := graph(accountID, conn)

// 	iNode, err := wrapFindNode(rPool, accountID, label, id)
// 	if err != nil {
// 		return nil, err
// 	}

// 	err = setProperties(&graph, iNode, newFields)

// 	Temp(rPool, accountID, label, quote("4d247443-b257-4b06-ba99-493cf9d83ce7"), id)

// 	return iNode, errors.Wrap(err, "err on commit")
// }

func wrapFindNode(rPool *redis.Pool, accountID, label, id string) (*rg.Node, error) {
	iNode, err := GetNode(rPool, accountID, label, id)
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

func Temp(rPool *redis.Pool, accountID, label, id string) (*rg.Node, error) {
	log.Printf("calling temp for %v", id)
	iNode := rg.Node{
		Label: label,
	}

	conn := rPool.Get()
	defer conn.Close()
	graph := graph(accountID, conn)

	query := fmt.Sprintf(`MATCH (i:%s) where i.id = "%s" return i`, quote(label), id)
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

// func setProperties(graph *rg.Graph, iNode *rg.Node, newFields []entity.Field) error {
// 	isModified := false
// 	graph.AddNode(iNode)

// 	for _, field := range newFields {
// 		if val, ok := iNode.Properties[field.Key]; (ok && field.Value == val) || field.Key == "id" {
// 			continue
// 		}
// 		isModified = true
// 		if !createEdges(graph, iNode, field) { // don't add properties if edges exists
// 			iNode.Properties[quote(field.Key)] = field.Value
// 		}
// 	}
// 	if isModified {
// 		_, err := Commit(graph)
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }

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

//GraphBluePrint takes an item and build the bule print of the nodes and relationships
type GraphBluePrint struct {
	GraphName  string
	Label      string
	ItemID     string
	Properties map[string]interface{} // default fields
	Contains   []GraphBluePrint       // list/map fields
	Has        []GraphBluePrint       // reference fields
}

//WhitelistedProperties filters the fields with types text area, lists, maps, and reference.
//It updates the properties with id and restrict field keys to updates the id.
func WhitelistedProperties(graphName, label, itemID string, fields []entity.Field) GraphBluePrint {
	baseGP := buildGraphBP(graphName, label).makeBaseGraphBp(itemID)
	for _, field := range fields {
		if field.IsKeyId() {
			continue
		}
		if field.List {
			list := field.Value.([]string)
			containsGraph := buildGraphBP(graphName, field.Key)
			for _, element := range list {
				log.Printf("element %v", element)
				baseGP = containsGraph.append(baseGP, element)
			}
		} else if field.DataType == entity.TypeReference {
		} else {
			baseGP.Properties[quote(field.Key)] = field.Value
		}

	}
	return baseGP
}

//GetNode fetches the node for the id provided
func GetNode(rPool *redis.Pool, graphName, label, itemID string) (*rg.Node, error) {
	conn := rPool.Get()
	defer conn.Close()
	graph := graph(graphName, conn)

	var iNode *rg.Node
	query := fmt.Sprintf(`MATCH (i:%s) where i.id = "%s" return i`, quote(label), itemID)
	result, err := graph.Query(query)
	if err != nil {
		return iNode, errors.Wrap(err, "selection nodes")
	}
	result.PrettyPrint()
	if result.Next() {
		records := result.Record().Values()
		if result.Next() || len(records) == 0 {
			return iNode, errors.New("fetching the node for id: " + itemID)
		}
		iNode = records[0].(*rg.Node)
		iNode.Label = label
	}
	return iNode, nil
}

//UpsertNode create/update the node with the given properties.
//Properties should not include text area, lists, maps, reference.
//TODO: For update, properties should include only the modified values including null for deleted keys.
func UpsertNode(rPool *redis.Pool, gbp GraphBluePrint) (*rg.QueryResult, error) {
	conn := rPool.Get()
	defer conn.Close()
	graph := graph(gbp.GraphName, conn)

	q := updateProperties(gbp)
	return graph.Query(q)
}

//"MERGE (n { id: '12345' }) SET n.age = 33, n.name = 'Bob'"
func updateProperties(gbp GraphBluePrint) string {
	alias := rg.RandomString(10)
	n := newNode(gbp.Label, alias, gbp.ItemID)
	s := mergeNode(n)
	if len(gbp.Properties) > 0 {
		p := make([]string, 0, len(gbp.Properties))
		for k, v := range gbp.Properties {
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

//UpsertEdge create/update the relationship between two nodes
func UpsertEdge(rPool *redis.Pool, gbp GraphBluePrint, label1 string) (*rg.QueryResult, error) {
	conn := rPool.Get()
	defer conn.Close()
	graph := graph(gbp.GraphName, conn)

	srcNode := newNode(gbp.Label, rg.RandomString(10), gbp.ItemID)
	dNode := &rg.Node{
		Alias: rg.RandomString(10),
		Label: quote(label1),
		Properties: map[string]interface{}{
			"element": "honda",
		},
	}

	q := updateRelationships("contains", srcNode, dNode)

	return graph.Query(q)
}

//"MERGE (charlie { name: 'Charlie Sheen' }) MERGE (wallStreet:Movie { name: 'Wall Street' }) MERGE (charlie)-[r:ACTED_IN]->(wallStreet)"
func updateRelationships(relation string, srcNode, destNode *rg.Node) string {
	edge := rg.EdgeNew(relation, srcNode, destNode, nil)
	ms := matchNode(srcNode)
	md := mergeNode(destNode)
	me := mergeEdge(edge)
	return strings.Join(ms, "") + strings.Join(md, "") + strings.Join(me, "")
}

func newNode(label, alias, id string) *rg.Node {
	n := rg.NodeNew(label, alias, map[string]interface{}{
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

func buildGraphBP(graphName, label string) GraphBluePrint {
	gbp := GraphBluePrint{
		GraphName:  graphName,
		Label:      quote(label),
		Properties: map[string]interface{}{},
	}
	return gbp
}

func (gbp GraphBluePrint) makeBaseGraphBp(itemID string) GraphBluePrint {
	gbp.ItemID = itemID
	gbp.Properties[quote(entity.FieldIdKey)] = gbp.ItemID
	gbp.Contains = make([]GraphBluePrint, 0)
	gbp.Has = make([]GraphBluePrint, 0)
	return gbp
}

func (cg GraphBluePrint) append(gbp GraphBluePrint, element interface{}) GraphBluePrint {
	cg.Properties[quote("element")] = element
	gbp.Contains = append(gbp.Contains, cg)
	return gbp
}
