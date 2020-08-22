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

var (
	// ErrNoEdgeNodesToAssociate will returned when there is no edge nodes are available for making the relation with source node
	ErrNoEdgeNodesToAssociate = errors.New("There is no edge nodes to associate with")
)

func graph(graphName string, conn redis.Conn) rg.Graph {
	return rg.GraphNew(graphName, conn)
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

//GraphNode takes an item and build the bule print of the nodes and relationships
type GraphNode struct {
	GraphName  string
	Label      string
	ItemID     string
	Properties map[string]PropValue // default fields
	Contains   []GraphNode          // list/map fields
	Has        []GraphNode          // reference fields
}

type PropValue struct {
	Operator string
	Value    interface{}
	Type     string
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
func UpsertNode(rPool *redis.Pool, gn GraphNode) (*rg.QueryResult, error) {
	conn := rPool.Get()
	defer conn.Close()
	graph := graph(gn.GraphName, conn)

	q := setNode(gn)
	log.Println("UpsertNodeQuery --> ", q)
	return graph.Query(q)
}

//"MERGE (n { id: '12345' }) SET n.age = 33, n.name = 'Bob'"
func setNode(gn GraphNode) string {
	srcNode := gn.RgNode(false)
	s := mergeNode(srcNode)
	if len(gn.Properties) > 0 {
		p := make([]string, 0, len(gn.Properties))
		for k, v := range gn.Properties {
			if k == "id" {
				continue
			}
			p = append(p, fmt.Sprintf("%s.%s = %v", srcNode.Alias, k, rg.ToString(v)))
		}

		s = append(s, "SET")
		s = append(s, strings.Join(p, ", "))
	}

	return strings.Join(s, " ")
}

//UpsertEdge creates/updates the relationship between src node and all its dst node
func UpsertEdge(rPool *redis.Pool, gn GraphNode) (*rg.QueryResult, error) {
	conn := rPool.Get()
	defer conn.Close()
	graph := graph(gn.GraphName, conn)

	srcNode := gn.RgNode(false)
	s := matchNode(srcNode)

	for _, cn := range gn.Contains {
		dstNode := cn.RgNode(true)
		s = append(s, mergeRelation("contains", srcNode, dstNode)...)
	}

	for _, hn := range gn.Has {
		dstNode := hn.RgNode(true)
		s = append(s, mergeRelation("has", srcNode, dstNode)...)
	}
	if len(gn.Contains) == 0 && len(gn.Has) == 0 {
		return nil, ErrNoEdgeNodesToAssociate
	}
	q := strings.Join(s, " ")
	log.Println("UpsertEdgeQuery --> ", q)
	return graph.Query(q)
}

//"MERGE (charlie { name: 'Charlie Sheen' }) MERGE (wallStreet:Movie { name: 'Wall Street' }) MERGE (charlie)-[r:ACTED_IN]->(wallStreet)"
func mergeRelation(relation string, srcNode, destNode *rg.Node) []string {
	edge := rg.EdgeNew(relation, srcNode, destNode, nil)
	md := mergeNode(destNode)
	me := mergeEdge(edge)
	return append(md, me...)
}

// MATCH   (cJeZNYvqro:`7d9c4f94-890b-484c-8189-91c3d7e8e50b`{`id`:"12345"})
func matchNode(n *rg.Node) []string {
	s := []string{"MATCH"}
	s = append(s, n.Encode())
	return s
}

func mergeNode(n *rg.Node) []string {
	s := []string{"MERGE"}
	s = append(s, n.Encode())
	return s
}

func matchEdge(e *rg.Edge) []string {
	s := []string{"MATCH"}
	s = append(s, e.Encode())
	return s
}

func mergeEdge(e *rg.Edge) []string {
	s := []string{"MERGE"}
	s = append(s, e.Encode())
	return s
}

func quote(label string) string {
	return fmt.Sprintf("`%s`", label)
}

func BuildGNode(graphName, label string) GraphNode {
	gn := GraphNode{
		GraphName:  graphName,
		Label:      quote(label),
		Properties: map[string]interface{}{},
		Contains:   make([]GraphNode, 0),
		Has:        make([]GraphNode, 0),
	}
	return gn
}

func (gn GraphNode) MakeBaseGNode(itemID string, fields []entity.Field) GraphNode {
	gn.ItemID = itemID
	gn.Properties[quote(entity.FieldIdKey)] = gn.ItemID

	for _, field := range fields {
		if field.IsKeyId() {
			continue
		}
		if field.List {
			for _, element := range field.Value.([]string) {
				cn := BuildGNode(gn.GraphName, field.Key)
				cn.Properties[quote("element")] = element
				gn.Contains = append(gn.Contains, cn)
			}
		} else if field.DataType == entity.TypeReference {
		} else {
			gn.Properties[quote(field.Key)] = field.Value
		}

	}
	return gn
}

func (gn GraphNode) RgNode(allProps bool) *rg.Node {
	if allProps {
		return rg.NodeNew(gn.Label, rg.RandomString(10), gn.Properties)
	}
	return rg.NodeNew(gn.Label, rg.RandomString(10), map[string]interface{}{
		quote(entity.FieldIdKey): gn.ItemID,
	})
}
