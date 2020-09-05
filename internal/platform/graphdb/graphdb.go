package graphdb

import (
	"fmt"
	"log"
	"strings"

	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
	rg "github.com/redislabs/redisgraph-go"
	"gitlab.com/vjsideprojects/relay/internal/entity"
)

//GraphDB creates the base GraphNode with the fields of an item.
//Once that has been created it uses base graphNode to create/update/segment/get in the redisGraph
//GraphNode handles two cases - 1) CRUD 2) Segmentation
//Both the use cases handled effortlessly by the graphNode.
//Make sure to understand some of the concepts as well, for example
//The lists field ite, will be added as the relationship with empty itemID
//The segmentation may/may not contain the itemID

var (
	// ErrNoEdgeNodesToAssociate will returned when there is no edge nodes are available for making the relation with source node
	ErrNoEdgeNodesToAssociate = errors.New("There is no edge nodes to associate with")
)

func graph(graphName string, conn redis.Conn) rg.Graph {
	return rg.GraphNew(graphName, conn)
}

//GraphNode takes an item and build the bule print of the nodes and relationships
type GraphNode struct {
	GraphName    string
	Label        string
	RelationName string
	ItemID       string
	Fields       []Field
	Relations    []GraphNode // list/map fields
	IsReverse    bool
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

	srcNode := gn.justNode()
	s := matchNode(srcNode)
	wc := where(gn, srcNode.Alias)

	for _, rn := range gn.Relations {
		dstNode := rn.justNode()
		edge := rg.EdgeNew(rn.RelationName, srcNode, dstNode, nil)
		if rn.IsReverse {
			edge = rg.EdgeNew(rn.RelationName, dstNode, srcNode, nil)
		}
		s = append(s, matchNode(dstNode)...)
		s = append(s, matchEdge(edge)...)
		wc = append(wc, where(rn, dstNode.Alias)...)
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
	p := make([]string, 0, len(gn.Fields))
	if len(gn.Fields) > 0 {
		for _, f := range gn.Fields {
			switch f.DataType {
			case TypeString:
				p = append(p, fmt.Sprintf("%s.%s %s \"%v\"", alias, f.Key, f.Expression, f.Value))
			case TypeNumber:
				p = append(p, fmt.Sprintf("%s.%s %s %v", alias, f.Key, f.Expression, f.Value))
			default:
				p = append(p, fmt.Sprintf("%s.%s %s %v", alias, f.Key, f.Expression, f.Value))
			}
		}
	}

	return p
}

//UpsertNode create/update the node with the given properties.
//Properties should not include text area, lists, maps, reference.
//TODO: For update, properties should include only the modified values including null for deleted keys.
//TODO: handle deleted field/field values
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
	srcNode := gn.rgNode()
	s := mergeNode(srcNode)
	props := gn.onlyProps()
	if len(props) > 0 {
		p := make([]string, 0, len(props))
		for k, v := range props {
			p = append(p, fmt.Sprintf("%s.%s = %v", srcNode.Alias, k, rg.ToString(v)))
		}
		s = append(s, "SET")
		s = append(s, strings.Join(p, ", "))
	}

	for _, rn := range gn.Relations {
		dstNode := rn.rgNode()
		s = append(s, mergeRelation(rn.RelationName, srcNode, dstNode)...)
	}

	return strings.Join(s, " ")
}

//UpsertEdge creates/updates the relationship between src node and all its dst node
func UpsertEdge(rPool *redis.Pool, gn GraphNode) (*rg.QueryResult, error) {
	conn := rPool.Get()
	defer conn.Close()
	graph := graph(gn.GraphName, conn)

	srcNode := gn.rgNode()
	s := matchNode(srcNode)

	for _, rn := range gn.Relations {
		dstNode := rn.rgNode()
		s = append(s, mergeRelation(rn.RelationName, srcNode, dstNode)...)
	}

	if len(gn.Relations) == 0 {
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

func BuildGNode(graphName, label string) GraphNode {
	gn := GraphNode{
		GraphName: graphName,
		Label:     quote(label),
		Fields:    []Field{},
		Relations: make([]GraphNode, 0),
	}
	return gn
}

func (gn GraphNode) MakeBaseGNode(itemID string, fields []Field) GraphNode {
	gn.ItemID = itemID

	for _, f := range fields {
		switch f.DataType {
		case entity.TypeList:
			for _, element := range f.Value.([]string) {
				f.Field.Value = element
				rn := BuildGNode(gn.GraphName, f.Key).
					MakeBaseGNode("", []Field{*f.Field}).relateLists()
				gn.Relations = append(gn.Relations, rn)
			}
		case entity.TypeReference:
			//TODO: handle cyclic looping
			for _, ref := range f.RefList() {
				rEntityID, rItemID := fetchRef(ref)
				rn := BuildGNode(gn.GraphName, rEntityID).
					MakeBaseGNode(rItemID, []Field{*f.Field}).relateRefs()
				gn.Relations = append(gn.Relations, rn)
			}
		default:
			gn.Fields = append(gn.Fields, f)
		}
	}
	return gn
}

//useful during all the upsertNode/upsertEdge
//when id is not empty, choose node with id alone. (ref/create/update use-case)
//when id is empty, choose all the properties. (list use-case)
func (gn GraphNode) rgNode() *rg.Node {
	if gn.ItemID == "" { //useful for list
		return rg.NodeNew(gn.Label, rg.RandomString(10), gn.onlyProps())
	}
	return gn.justNode()
}

//useful during the get/upsertNode/upsertEdge
//when id is not empty - form node with id alone. (ref/create/update use-case)
//when id is empty - form node with out properties (segment use-case)
func (gn GraphNode) justNode() *rg.Node {
	properties := map[string]interface{}{}
	if gn.ItemID != "" { //useful for almost all the cases except the segmentation use-case
		properties[quote(FieldIdKey)] = gn.ItemID
	}
	return rg.NodeNew(gn.Label, rg.RandomString(10), properties)
}

//useful during the upsert.
func (gn GraphNode) onlyProps() map[string]interface{} {
	properties := map[string]interface{}{}
	for _, field := range gn.Fields {
		properties[quote(field.Key)] = field.Value
	}
	return properties
}

func (gn GraphNode) relateRefs() GraphNode {
	gn.RelationName = "has"
	if gn.ItemID == "" { // reverse true on segmentations ref
		gn.IsReverse = true
	}
	return gn
}

func (gn GraphNode) relateLists() GraphNode {
	gn.RelationName = "contains"
	return gn
}

func quote(label string) string {
	return fmt.Sprintf("`%s`", label)
}
