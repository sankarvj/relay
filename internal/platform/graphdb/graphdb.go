package graphdb

import (
	"encoding/json"
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
	Unlink       bool
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
	wi, wh := where(gn, srcNode.Alias, srcNode.Alias)
	s = append(s, wi...)
	if len(wh) > 0 {
		s = append(s, "WHERE")
		s = append(s, strings.Join(wh, " AND "))
	}

	for _, rn := range gn.Relations {
		dstNode := rn.justNode()
		edge := rg.EdgeNew(rn.RelationName, srcNode, dstNode, nil)
		if rn.IsReverse {
			edge = rg.EdgeNew(rn.RelationName, dstNode, srcNode, nil)
		}
		s = append(s, matchNode(dstNode)...)
		s = append(s, matchEdge(edge)...)
		wi, wh := where(rn, dstNode.Alias, srcNode.Alias)
		s = append(s, wi...)
		s = append(s, "WHERE")
		s = append(s, strings.Join(wh, " AND "))
	}

	q := strings.Join(s, " ")
	q = fmt.Sprintf("%s %s", q, fmt.Sprintf("RETURN %s", srcNode.Alias))

	result, err := graph.Query(q)
	log.Println("GetResultQuery gn.GraphName--> ", gn.GraphName)
	log.Println("GetResultQuery result--> ", q)
	log.Println("GetResultQuery err--> ", err)
	if err != nil {
		return result, err
	}
	result.PrettyPrint()
	return result, err
}

func where(gn GraphNode, alias, srcAlias string) ([]string, []string) {
	wi := make([]string, 0)
	wh := make([]string, 0, len(gn.Fields))
	if len(gn.Fields) > 0 {
		for _, f := range gn.Fields {

			if f.Aggr != "" {
				f.WithAlias = fmt.Sprintf("%s_%s", f.Aggr, f.Key)
				with := fmt.Sprintf("WITH %s,%s(%s.%s) as %s", srcAlias, f.Aggr, alias, f.Key, f.WithAlias)
				wi = append(wi, with)
			} else {
				f.WithAlias = fmt.Sprintf("%s.%s", alias, f.Key)
			}

			switch f.DataType {
			case TypeString:
				wh = append(wh, fmt.Sprintf("%s %s \"%v\"", f.WithAlias, f.Expression, f.Value))
			case TypeNumber:
				wh = append(wh, fmt.Sprintf("%s %s %v", f.WithAlias, f.Expression, f.Value))
			default:
				wh = append(wh, fmt.Sprintf("%s %s %v", f.WithAlias, f.Expression, f.Value))
			}
		}
	}

	return wi, wh
}

//UpsertNode create/update the node with the given properties.
//Properties should not include text area, lists, maps, reference.
//TODO: For update, properties should include only the modified values including null for deleted keys.
//TODO: handle deleted field/field values
func UpsertNode(rPool *redis.Pool, gn GraphNode) error {
	conn := rPool.Get()
	defer conn.Close()
	graph := graph(gn.GraphName, conn)

	srcNode := gn.rgNode()

	//first part set and new edge
	mps := mergeProperties(gn, srcNode)
	rs, ru := updateRelation(gn, srcNode)
	if len(mps) > 0 || len(rs) > 0 {
		s := mergeNode(srcNode)
		s = append(s, mps...)
		s = append(s, rs...)
		sq := strings.Join(s, " ")
		log.Println("UpsertNodeQuery --> ", sq)
		_, err := graph.Query(sq)
		if err != nil {
			return err
		}
	}

	//second part unlink edge
	//TODO: improvise this query. Instead of hitting n+1 times batch delete
	//Try like this: match a,b,c delete a,b,c
	for _, ruQ := range ru {
		s := matchNode(srcNode)
		s = append(s, ruQ)
		sq := strings.Join(s, " ")
		log.Println("UnlinkNodeQuery --> ", sq)
		_, err := graph.Query(sq)
		if err != nil {
			return err
		}
	}

	return nil
}

//"MERGE (n { id: '12345' }) SET n.age = 33, n.name = 'Bob'"
func mergeProperties(gn GraphNode, srcNode *rg.Node) []string {
	s := []string{}
	props := gn.onlyProps()
	if len(props) > 0 {
		p := make([]string, 0, len(props))
		for k, v := range props {
			p = append(p, fmt.Sprintf("%s.%s = %v", srcNode.Alias, k, rg.ToString(v)))
		}
		s = append(s, "SET")
		s = append(s, strings.Join(p, ", "))
	}
	return s
}

//UpsertEdge creates/updates the relationship between src node and all its dst node
func UpsertEdge(rPool *redis.Pool, gn GraphNode) error {
	conn := rPool.Get()
	defer conn.Close()
	graph := graph(gn.GraphName, conn)

	srcNode := gn.rgNode()
	rs, ru := updateRelation(gn, srcNode)

	//unlink existing relations
	for _, ulink := range ru {
		s := matchNode(srcNode)
		s = append(s, ulink)
		ruq := strings.Join(s, " ")
		log.Println("UnlinkEdgeQuery --> ", ruq)
		_, err := graph.Query(ruq)
		if err != nil {
			return err
		}
	}

	//make new relations
	if len(rs) > 0 {
		s := matchNode(srcNode)
		s = append(s, rs...)
		rsq := strings.Join(s, " ")
		log.Println("UpsertEdgeQuery --> ", rsq)
		_, err := graph.Query(rsq)
		if err != nil {
			return err
		}
	}
	return nil
}

func updateRelation(gn GraphNode, srcNode *rg.Node) ([]string, []string) {
	s, u := []string{}, []string{}
	for _, rn := range gn.Relations {
		dstNode := rn.rgNode()
		if rn.Unlink {
			u = append(u, strings.Join(unlinkRelation(rn.RelationName, srcNode, dstNode), " "))
		} else {
			s = append(s, mergeRelation(rn.RelationName, srcNode, dstNode)...)
		}
	}
	return s, u
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

//"MERGE (charlie { name: 'Charlie Sheen' }) MATCH (wallStreet:Movie { name: 'Wall Street' }) MATCH (charlie)-[r:ACTED_IN]->(wallStreet)"
func unlinkRelation(relation string, srcNode, destNode *rg.Node) []string {
	unlinkAlias := rg.RandomString(10)
	edge := rg.EdgeNew(relation, srcNode, destNode, nil)
	md := matchNode(destNode)
	me := unlinkEdge(unlinkAlias, edge)
	return append(md, me...)
}

func unlinkEdge(unlinkAlias string, e *rg.Edge) []string {
	s := []string{"MATCH"}
	s = append(s, deleteEncode(e, unlinkAlias))
	s = append(s, "DELETE", unlinkAlias)
	return s
}

func BuildGNode(graphName, label string, unlink bool) GraphNode {
	gn := GraphNode{
		GraphName: graphName,
		Label:     label,
		Fields:    []Field{},
		Relations: make([]GraphNode, 0),
		Unlink:    unlink,
	}
	return gn
}

func (gn GraphNode) MakeBaseGNode(itemID string, fields []Field) GraphNode {
	gn.ItemID = itemID

	for _, f := range fields {
		switch f.DataType {
		case entity.TypeList:
			for i, element := range f.Value.([]interface{}) {
				f.Field.Value = element
				rn := BuildGNode(gn.GraphName, f.Key, f.doUnlink(i)).
					MakeBaseGNode("", []Field{*f.Field}).relateLists()
				gn.Relations = append(gn.Relations, rn)
			}
		case entity.TypeReference:
			//TODO: handle cyclic looping
			if f.Value == nil {
				continue
			}
			for i, rItemID := range f.Value.([]interface{}) {
				rEntityID := f.RefID
				rn := BuildGNode(gn.GraphName, rEntityID, f.doUnlink(i)).
					MakeBaseGNode(rItemID.(string), []Field{*f.Field}).relateRefs(f.IsReverse)
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
		return rg.NodeNew(quote(gn.Label), rg.RandomString(10), gn.onlyProps())
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
	return rg.NodeNew(quote(gn.Label), rg.RandomString(10), properties)
}

//useful during the upsert.
func (gn GraphNode) onlyProps() map[string]interface{} {
	properties := map[string]interface{}{}
	for _, field := range gn.Fields {
		properties[quote(field.Key)] = field.Value
	}
	return properties
}

func (gn GraphNode) relateRefs(reverse bool) GraphNode {
	gn.RelationName = "has"
	if reverse { // reverse true on segmentations when segmenting contacts which are associated to deals. (deals entity has contacts but contacts entity do not have the relationship with deals)
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

func (gn GraphNode) JsonB() (string, error) {
	jsonbody, err := json.Marshal(gn)
	if err != nil {
		return "", err
	}
	return string(jsonbody), err
}

func (gn GraphNode) AddIDCondition(itemID interface{}) GraphNode {
	f := Field{
		Expression: "=",
		Key:        FieldIdKey,
		DataType:   TypeString,
		Value:      itemID,
	}
	gn.Fields = append(gn.Fields, f)
	return gn
}

func GraphNodeSt(jsonB string) (GraphNode, error) {
	var gn GraphNode
	if err := json.Unmarshal([]byte(jsonB), &gn); err != nil {
		return gn, errors.Wrapf(err, "error while unmarshalling graph node")
	}
	return gn, nil
}

//deleteEncode adds the alias for relation.
//current go-library doesn't have this functionality
//overloading encode with relation alias
func deleteEncode(e *rg.Edge, relationAlias string) string {
	s := []string{"(", e.Source.Alias, ")"}

	s = append(s, "-[")

	if e.Relation != "" {
		s = append(s, relationAlias, ":", e.Relation) // this is the only change from the source edge.go
	}

	if len(e.Properties) > 0 {
		p := make([]string, 0, len(e.Properties))
		for k, v := range e.Properties {
			p = append(p, fmt.Sprintf("%s:%v", k, rg.ToString(v)))
		}

		s = append(s, "{")
		s = append(s, strings.Join(p, ","))
		s = append(s, "}")
	}

	s = append(s, "]->")
	s = append(s, "(", e.Destination.Alias, ")")

	return strings.Join(s, "")
}
