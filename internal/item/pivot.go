package item

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
	rg "github.com/redislabs/redisgraph-go"
	"gitlab.com/vjsideprojects/relay/internal/entity"
	"gitlab.com/vjsideprojects/relay/internal/platform/segment"
)

var (
	// ErrNoEdgeNodesToAssociate will returned when there is no edge nodes are available for making the relation with source node
	ErrNoEdgeNodesToAssociate = errors.New("There is no edge nodes to associate with")
)

func graph(graphName string, conn redis.Conn) rg.Graph {
	return rg.GraphNew(graphName, conn)
}

//GraphNode takes an item and build the bule print of the nodes and relationships
type GraphNode struct {
	GraphName  string
	Label      string
	ItemID     string
	Properties map[string]interface{}       // default fields
	Condition  map[string]segment.Condition // meta field attr
	Contains   []GraphNode                  // list/map fields
	Has        []GraphNode                  // reference fields
}

type PropValue struct {
	Operator string
	Type     string
	Key      string
	Value    interface{}
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

//GetResult fetches resultant node
func GetResult(rPool *redis.Pool, gn GraphNode) (*rg.QueryResult, error) {
	conn := rPool.Get()
	defer conn.Close()
	graph := graph(gn.GraphName, conn)

	srcNode := gn.RgNode(false)
	s := matchNode(srcNode)
	wc := where(gn, srcNode.Alias)

	for _, cn := range gn.Contains {
		dstNode := cn.RgNode(false)
		edge := rg.EdgeNew("contains", srcNode, dstNode, nil)
		s = append(s, matchNode(dstNode)...)
		s = append(s, matchEdge(edge)...)
		wc = append(wc, where(cn, dstNode.Alias)...)
	}

	q := strings.Join(s, " ")
	w := strings.Join(wc, " AND ")
	q = strings.Join([]string{q, w}, " WHERE ")
	q = fmt.Sprintf("%s %s", q, fmt.Sprintf("RETURN %s", srcNode.Alias))

	result, err := graph.Query(q)
	log.Println("GetResultQuery --> ", q)
	result.PrettyPrint()
	return result, err
}

func where(gn GraphNode, alias string) []string {
	p := make([]string, 0, len(gn.Properties))
	if len(gn.Condition) > 0 {
		for _, condition := range gn.Condition {
			if condition.Type == "S" {
				p = append(p, fmt.Sprintf("%s.%s %s \"%v\"", alias, condition.Key, condition.Operator, condition.Value))
			} else if condition.Type == "N" {
				p = append(p, fmt.Sprintf("%s.%s %s %v", alias, condition.Key, condition.Operator, condition.Value))
			} else {
				p = append(p, fmt.Sprintf("%s.%s %s %v", alias, condition.Key, condition.Operator, condition.Value))
			}
		}
	}

	return p
}

//UpsertNode create/update the node with the given properties.
//Properties should not include text area, lists, maps, reference.
//TODO: For update, properties should include only the modified values including null for deleted keys.
func UpsertNode(rPool *redis.Pool, gn GraphNode) (*rg.QueryResult, error) {
	conn := rPool.Get()
	defer conn.Close()
	graph := graph(gn.GraphName, conn)

	q := mergeProperties(gn)
	log.Println("UpsertNodeQuery --> ", q)
	return graph.Query(q)
}

//"MERGE (n { id: '12345' }) SET n.age = 33, n.name = 'Bob'"
func mergeProperties(gn GraphNode) string {
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
		Condition:  map[string]segment.Condition{},
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
		if field.DataType == entity.TypeList {
			for _, element := range field.Value.([]string) {
				cn := BuildGNode(gn.GraphName, field.Key)
				cn.Properties[field.Field.Key] = element
				gn.Contains = append(gn.Contains, cn)
			}
		} else if field.DataType == entity.TypeReference {
		} else {
			gn.Properties[quote(field.Key)] = field.Value
		}

	}
	return gn
}

func (gn GraphNode) SegmentBaseGNode(seg segment.Segment) GraphNode {
	for i, condition := range seg.Conditions {
		if condition.Type == "L" {
			cn := BuildGNode(gn.GraphName, condition.Key)
			cn.Condition[strconv.Itoa(i)] = *condition.Condition
			gn.Contains = append(gn.Contains, cn)
		} else if condition.Type == "R" {

		} else {
			gn.Condition[strconv.Itoa(i)] = condition
		}

	}
	return gn
}

func (gn GraphNode) RgNode(withProps bool) *rg.Node {
	if withProps {
		return rg.NodeNew(gn.Label, rg.RandomString(10), gn.Properties)
	}
	if gn.ItemID != "" {
		return rg.NodeNew(gn.Label, rg.RandomString(10), map[string]interface{}{
			quote(entity.FieldIdKey): gn.ItemID,
		})
	}
	return rg.NodeNew(gn.Label, rg.RandomString(10), nil)
}
