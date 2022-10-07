package graphdb

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
	rg "github.com/redislabs/redisgraph-go"
	"gitlab.com/vjsideprojects/relay/internal/platform/ruleengine/services/lexer/lexertoken"
	"gitlab.com/vjsideprojects/relay/internal/platform/util"
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
	GraphName     string
	Label         string
	RelationName  string
	ItemID        string
	Fields        []Field
	Relations     []GraphNode // list/map fields
	SourceNode    *rg.Node    // not a good idea. used for count
	ReturnNode    *rg.Node    // not a good idea. used for count
	UseReturnNode bool        // not a good idea. used to return dstnode on searcing list fields
	unlink        bool
	isReverse     bool
	Optional      bool
}

func BuildGNode(graphName, label string, unlink bool) GraphNode {
	gn := GraphNode{
		GraphName: graphName,
		Label:     label,
		Fields:    []Field{},
		Relations: make([]GraphNode, 0),
		unlink:    unlink,
	}
	return gn
}

func (gn GraphNode) MakeBaseGNode(itemID string, fields []Field) GraphNode {
	gn.ItemID = itemID

	for _, f := range fields {
		if f.Value == nil {
			continue
		}
		switch f.DataType {
		case TypeNumber:
			f.Value = util.ConvertToNumber(f.Value)
			gn.Fields = append(gn.Fields, f)
		case TypeList:
			for i, element := range f.Value.([]interface{}) {
				var value string
				switch v := element.(type) {
				case int:
					value = util.ConvertIntToStr(v)
				case string:
					value = v
				}
				rn := BuildGNode(gn.GraphName, f.Key, f.doUnlink(i)).
					MakeBaseGNode(value, []Field{*f.Field}).relateLists()
				gn.Relations = append(gn.Relations, rn)
			}
		case TypeReference:
			//TODO: handle cyclic looping
			for i, rItemID := range f.Value.([]interface{}) {
				rEntityID := f.RefID
				rn := BuildGNode(gn.GraphName, rEntityID, f.doUnlink(i)).
					MakeBaseGNode(rItemID.(string), []Field{*f.Field}).relateRefs(f.IsReverse)
				gn.Relations = append(gn.Relations, rn)
			}

			// if len(f.Value.([]interface{})) == 0 { //TODO: hacky-none-fix block
			// 	rEntityID := f.RefID
			// 	rn := BuildGNode(gn.GraphName, rEntityID, false).MakeBaseGNode("", []Field{*f.Field}).relateRefs(f.IsReverse)
			// 	gn.Relations = append(gn.Relations, rn)
			// }

		case TypeDateTime, TypeDate: // converts the time to timestamp during upsert for easy filtering.
			if f.Value != nil && f.Value != "" {
				t, err := util.ParseTime(f.Value.(string))
				if err != nil {
					log.Println("***> unexpected/unhandled error occurred. Unbale to convert the datetime str. Please fix the value ", f.Value)
				}
				f.Value = util.GetMilliSecondsFloat(t)
			}
			gn.Fields = append(gn.Fields, f)
		case TypeDateRange, TypeDateTimeMillis: // using the time range, time in millis formats calculated in the base.go -> makeGraphField
			gn.Fields = append(gn.Fields, f)
		default:
			gn.Fields = append(gn.Fields, f)
		}
	}
	return gn
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
func GetResult(rPool *redis.Pool, gn GraphNode, pageNo int, sortBy, direction string) (*rg.QueryResult, error) {
	conn := rPool.Get()
	defer conn.Close()

	graph := graph(gn.GraphName, conn)
	q := makeQuery(rPool, &gn)

	returnAlias := gn.SourceNode.Alias
	if gn.UseReturnNode {
		returnAlias = gn.ReturnNode.Alias
	}
	q = fmt.Sprintf("%s %s", q, fmt.Sprintf("RETURN %s", returnAlias))

	//sorting
	if sortBy != "" && direction != "" {
		q = fmt.Sprintf("%s %s", q, fmt.Sprintf("ORDER BY %s.`%s` %s", returnAlias, sortBy, direction))
	}
	skipCount := pageNo * util.PageLimt
	//pagination
	q = fmt.Sprintf("%s %s", q, fmt.Sprintf("SKIP %d LIMIT %d", skipCount, util.PageLimt))

	result, err := graph.Query(q)
	if err != nil {
		//DEBUG LOG
		log.Printf("*********> debug: internal.platform.graphdb : graphdb - query: %s - err:%v\n", q, err)
		return result, err
	}
	//DEBUG LOG
	log.Printf("*********> debug: internal.platform.graphdb : graphdb - result query: %s\n", q)
	//DEBUG LOG log.Printf("*********> debug: internal.platform.graphdb : graphdb - result: %v\n", result)
	//result.PrettyPrint()
	return result, err
}

//GetCount fetches count of destination node
func GetCount(rPool *redis.Pool, gn GraphNode, groupById bool) (*rg.QueryResult, error) {
	conn := rPool.Get()
	defer conn.Close()
	graph := graph(gn.GraphName, conn)

	q := makeQuery(rPool, &gn)

	if groupById {
		q = fmt.Sprintf("%s %s", q, fmt.Sprintf("RETURN COUNT(%s), %s.id", gn.SourceNode.Alias, gn.ReturnNode.Alias))
	} else {
		q = fmt.Sprintf("%s %s", q, fmt.Sprintf("RETURN COUNT(%s)", gn.SourceNode.Alias))
	}

	result, err := graph.Query(q)
	if err != nil {
		//DEBUG LOG
		log.Printf("*********> debug: internal.platform.graphdb : graphdb - count query: %s - err:%v\n", q, err)
		return result, err
	}
	//DEBUG LOG
	//log.Printf("*********> debug: internal.platform.graphdb : graphdb - count query: %s\n", q)
	//log.Printf("*********> debug: internal.platform.graphdb : graphdb - count result: %v\n", result)
	result.PrettyPrint()
	return result, err
}

func GetFromParentCount(rPool *redis.Pool, gn GraphNode) (*rg.QueryResult, error) {
	conn := rPool.Get()
	defer conn.Close()
	graph := graph(gn.GraphName, conn)

	q := makeQuery(rPool, &gn)

	q = fmt.Sprintf("%s %s", q, fmt.Sprintf("RETURN COUNT(distinct %s)", gn.ReturnNode.Alias))

	result, err := graph.Query(q)
	if err != nil {
		//DEBUG LOG
		log.Printf("*********> debug: internal.platform.graphdb : graphdb - count query: %s - err:%v\n", q, err)
		return result, err
	}
	//DEBUG LOG
	//log.Printf("*********> debug: internal.platform.graphdb : graphdb - count query: %s\n", q)
	//log.Printf("*********> debug: internal.platform.graphdb : graphdb - count result: %v\n", result)
	result.PrettyPrint()
	return result, err
}

func GetSum(rPool *redis.Pool, gn GraphNode, sumKey string) (*rg.QueryResult, error) {
	conn := rPool.Get()
	defer conn.Close()
	graph := graph(gn.GraphName, conn)

	q := makeQuery(rPool, &gn)

	q = fmt.Sprintf("%s %s", q, fmt.Sprintf("RETURN SUM(%s.`%s`)", gn.SourceNode.Alias, sumKey))

	result, err := graph.Query(q)
	if err != nil {
		//DEBUG LOG
		log.Printf("*********> debug: internal.platform.graphdb : graphdb - sum query: %s - err:%v\n", q, err)
		return result, err
	}
	//DEBUG LOG
	//log.Printf("*********> debug: internal.platform.graphdb : graphdb - result: %v\n", result)

	result.PrettyPrint()
	return result, err
}

func GetGroupedCount(rPool *redis.Pool, gn GraphNode, groupById string) (*rg.QueryResult, error) {
	conn := rPool.Get()
	defer conn.Close()
	graph := graph(gn.GraphName, conn)

	q := makeQuery(rPool, &gn)

	q = fmt.Sprintf("%s %s", q, fmt.Sprintf("RETURN COUNT(%s), %s.`%s`", gn.SourceNode.Alias, gn.SourceNode.Alias, groupById))

	result, err := graph.Query(q)
	if err != nil {
		//DEBUG LOG
		log.Printf("*********> debug: internal.platform.graphdb : graphdb - count query: %s - err:%v\n", q, err)
		return result, err
	}
	//DEBUG LOG
	//log.Printf("*********> debug: internal.platform.graphdb : graphdb - count query: %s\n", q)
	//log.Printf("*********> debug: internal.platform.graphdb : graphdb - count result: %v\n", result)
	result.PrettyPrint()
	return result, err
}

func Delete(rPool *redis.Pool, graphName, label, itemID string) error {
	conn := rPool.Get()
	defer conn.Close()
	graph := graph(graphName, conn)

	query := fmt.Sprintf(`MATCH (i:%s) where i.id = "%s" DELETE i`, quote(label), itemID)
	_, err := graph.Query(query)
	if err != nil {
		return errors.Wrap(err, "deleting nodes")
	}

	return nil
}

func makeQuery(rPool *redis.Pool, gn *GraphNode) string {
	srcNode := gn.justNode()
	gn.SourceNode = srcNode
	s := matchNode(srcNode)
	wi, wh := where(gn, srcNode.Alias, srcNode.Alias)

	s = append(s, wi...)
	if len(wh) > 0 {
		s = append(s, "WHERE")
		s = append(s, strings.Join(wh, " AND "))
	}

	s = chainRelations(gn, srcNode, s)
	q := strings.Join(s, " ")
	return q
}

func where(gn *GraphNode, alias, srcAlias string) ([]string, []string) {
	wi := make([]string, 0)
	wh := make([]string, 0, len(gn.Fields))
	if len(gn.Fields) > 0 {
		for _, f := range gn.Fields {

			if f.Aggr != "" {
				f.WithAlias = fmt.Sprintf("%s_%s", f.Aggr, f.Key)
				with := fmt.Sprintf("WITH %s,%s(%s.`%s`) as %s", srcAlias, f.Aggr, alias, f.Key, f.WithAlias)
				wi = append(wi, with)
			} else {
				f.WithAlias = fmt.Sprintf("%s.`%s`", alias, f.Key)
			}

			switch f.Expression {
			case operatorMap[lexertoken.LikeSign]:
				f.Value = strings.ToLower(fmt.Sprintf("%v", f.Value))
				f.WithAlias = fmt.Sprintf("tolower(%s)", f.WithAlias)
			case operatorMap[lexertoken.NotINSign]: //to support `WHERE NOT qvHZjOKbzM.`id` IN ["0ce398f5-8d85-4436-af0f-b884d18ecc5a"]`
				f.Expression = lexertoken.INSign
				f.WithAlias = fmt.Sprintf("NOT %s", f.WithAlias)
				gn.Optional = true
			}

			switch f.DataType {
			case TypeString:
				wh = append(wh, fmt.Sprintf("%s %s \"%v\"", f.WithAlias, f.Expression, f.Value))
			case TypeNumber:
				wh = append(wh, fmt.Sprintf("%s %s %v", f.WithAlias, f.Expression, f.Value))
			case TypeDateTime: //datetime in graph DB always expects a range
				wh = append(wh, fmt.Sprintf("%s %s %v", f.WithAlias, f.Expression, f.Value))
			case TypeDateRange:
				wh = append(wh, fmt.Sprintf("%s %s %v AND %s %s %v", f.WithAlias, ">", f.Min, f.WithAlias, "<", f.Max))
			case TypeWist:
				wh = append(wh, fmt.Sprintf("%s %s %v", f.WithAlias, f.Expression, quoteSlice(f.Value.([]string))))
			default:
				wh = append(wh, fmt.Sprintf("%s %s %v", f.WithAlias, f.Expression, f.Value))
			}
		}
	}

	return wi, wh
}

func chainRelations(gn *GraphNode, srcNode *rg.Node, s []string) []string {
	for _, rn := range gn.Relations {
		dstNode := rn.justNode()
		gn.ReturnNode = dstNode
		var localS []string
		if len(rn.Relations) > 0 {
			localS = chainRelations(&rn, dstNode, localS)
		}

		edge := rg.EdgeNew(rn.RelationName, srcNode, dstNode, nil)
		if rn.isReverse {
			edge = rg.EdgeNew(rn.RelationName, dstNode, srcNode, nil)
		}

		wi, wh := where(&rn, dstNode.Alias, srcNode.Alias)

		//hacky-fix-none
		s = append(s, matchNode(dstNode)...)
		if rn.Optional {
			s = append(s, optionalMatchEdge(edge)...)
		} else {
			s = append(s, matchEdge(edge)...)
		}

		s = append(s, wi...)
		if len(wh) > 0 {
			s = append(s, "WHERE")
			s = append(s, strings.Join(wh, " AND "))
		}

		s = append(s, localS...)

	}
	return s
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
		//DEBUGGING LOG
		//log.Println("internal.platform.graphdb upsert edge query:", sq)
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
		log.Println("internal.platform.graphdb unlink edge query:", sq)
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
		//DEBUGGING LOG log.Println("internal.platform.graphdb unlink edge query:", ruq)
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
		//DEBUGGING LOG log.Println("internal.platform.graphdb upsert edge query:", rsq)
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
		if rn.unlink {
			u = append(u, strings.Join(unlinkRelation(rn.RelationName, srcNode, dstNode), " "))
		} else {
			if rn.isReverse {
				s = append(s, mergeRevRelation(rn.RelationName, srcNode, dstNode)...)
			} else {
				s = append(s, mergeRelation(rn.RelationName, srcNode, dstNode)...)
			}

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

//Usefull when we update the event props with parent edge
func mergeRevRelation(relation string, srcNode, destNode *rg.Node) []string {
	edge := rg.EdgeNew(relation, destNode, srcNode, nil)
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

func optionalMatchEdge(e *rg.Edge) []string {
	s := []string{"OPTIONAL MATCH"}
	s = append(s, e.Encode())
	s = append(s, fmt.Sprintf("WITH %s ", e.Source.Alias))
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

func (gn GraphNode) ParentEdge(parentEntityID, parentItemID string, rev bool) GraphNode {
	rn := BuildGNode(gn.GraphName, parentEntityID, false).
		MakeBaseGNode(parentItemID, []Field{}).relateRefs(rev)
	gn.Relations = append(gn.Relations, rn)
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
		gn.isReverse = true
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

func quoteSlice(strs []string) string {
	commaSep := "\"" + strings.Join(strs, "\", \"") + "\""
	return fmt.Sprintf("[%s]", commaSep)
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

var operatorMap = map[string]string{
	lexertoken.EqualSign:    "=",
	lexertoken.NotEqualSign: "!=",
	lexertoken.GTSign:       ">",
	lexertoken.LTSign:       "<",
	lexertoken.LikeSign:     "STARTS WITH",
	lexertoken.INSign:       "IN",
	lexertoken.NotINSign:    "NOT IN",
	lexertoken.BFSign:       "<",
	lexertoken.AFSign:       ">",
}

//TODO genralise and remove the if check
//In segmentation call sometimes the value will be passed as 'eq' and sometime '='
func Operator(lexerOp string) string {
	if val, ok := operatorMap[lexerOp]; ok {
		return val
	}
	return lexerOp
}
