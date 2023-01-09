package util

import (
	"encoding/json"
	"expvar"
	"fmt"
	"log"
	"net/mail"
	"strconv"
	"strings"

	pluralize "github.com/gertd/go-pluralize"
	rg "github.com/redislabs/redisgraph-go"
)

const (
	PageLimt int = 5
	MaxLimt  int = 500
)

func ConvertSliceType(s []string) []interface{} {
	inf := make([]interface{}, len(s))
	for i, v := range s {
		inf[i] = v
	}
	return inf
}

func ConvertSliceTypeRev(inf []interface{}) []string {
	s := make([]string, len(inf))
	for i, v := range inf {
		s[i] = fmt.Sprint(v)
	}
	return s
}

func ConvertStrToPtStr(inf []string) []*string {
	s := make([]*string, len(inf))
	for i, v := range inf {
		s[i] = &v
	}
	return s
}

func ConvertInterfaceToMap(intf interface{}) map[string]interface{} {
	var itemMap map[string]interface{}
	jsonbody, err := json.Marshal(intf)
	if err != nil {
		log.Println("*> expected error occured when marshalling on `ConvertInterfaceToMap` sending empty map. error:", err)
		return itemMap
	}
	err = json.Unmarshal(jsonbody, &itemMap)
	if err != nil {
		log.Println("*> expected error occured when unmarshalling on `ConvertInterfaceToMap` sending empty map. error:", err)
	}
	return itemMap
}

func ConvertIntToStr(i int) string {
	return strconv.FormatInt(int64(i), 10)
}

func ConvertStrToInt(s string) int {
	if i, err := strconv.Atoi(s); err == nil {
		return i
	}
	log.Printf("*> expected error occured when converting string %s to int on `ConvertStrToInt` sending zero\n", s)
	return 0
}

func ConvertStrToInt64(s string) int64 {
	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		return i
	}
	log.Printf("expected error occured when converting string %s to int64 on `ConvertStrToInt64` sending zero\n", s)
	return 0
}

func ConvertToNumber(in interface{}) interface{} {
	switch v := in.(type) {
	default:
		return in
	case int, int64, float64:
		return v
	case string:
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return 0
		}
		return f
	}
}

func ConvertIntfToCommaSepString(in interface{}) string {
	switch v := in.(type) {
	default:
		return in.(string)
	case int, int64, float64:
		return ConvertIntToStr(v.(int))
	case []interface{}:
		strArr := make([]string, 0)
		for _, ev := range v {
			strArr = append(strArr, ev.(string))
		}
		return strings.Join(strArr[:], ",")
	case []string:
		strArr := make([]string, 0)
		for _, ev := range v {
			strArr = append(strArr, fmt.Sprintf("'%s'", ev))
		}
		return strings.Join(strArr[:], ",")
	}
}

func ConvertIntfToStr(intf interface{}) string {
	if intf != nil {
		return intf.(string)
	}
	return ""
}

func Contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func SubDomain(email string) string {
	at := strings.LastIndex(email, "@")
	if at >= 0 {
		_, domain := email[:at], email[at+1:]
		components := strings.Split(domain, ".")
		return components[0]
	} else {
		fmt.Printf("Error: %s is an invalid email address\n", email)
		return ""
	}
}

func MessageID(reference string) string {
	at := strings.LastIndex(reference, "@")
	if at >= 0 {
		messageID, _ := reference[:at], reference[at+1:]
		return messageID
	} else {
		fmt.Printf("Error: %s is an invalid reference address\n", reference)
		return ""
	}
}

func NameInEmail(email string) string {
	at := strings.LastIndex(email, "@")
	if at >= 0 {
		return email[:at]
	}
	return ""
}

func MainMailRefernce(reference string) string {
	components := strings.Split(reference, " ")
	if len(components) >= 0 {
		return components[0]
	} else {
		fmt.Printf("Error: %s is an invalid reference value\n", components)
		return ""
	}
}

func AddExpression(exTexp string, newExp string) string {
	if exTexp == "" {
		return newExp
	}
	if newExp == "" {
		return exTexp
	}
	return fmt.Sprintf("%s && %s", exTexp, newExp)
}

func Similar(a, b []interface{}) []interface{} {
	mb := make(map[interface{}]bool, len(b))
	for _, x := range b {
		mb[x] = true
	}
	var same []interface{}
	for _, x := range a {
		if _, found := mb[x]; found {
			same = append(same, x)
		}
	}
	return same
}

func Differ(a, b []interface{}) []interface{} {
	mb := make(map[interface{}]bool, len(b))
	for _, x := range b {
		mb[x] = true
	}
	var differ []interface{}
	for _, x := range a {
		if _, found := mb[x]; !found {
			differ = append(differ, x)
		}
	}
	return differ
}

func Similars(a, b []string) []string {
	mb := make(map[string]bool, len(b))
	for _, x := range b {
		mb[x] = true
	}
	var same []string
	for _, x := range a {
		if _, found := mb[x]; found {
			same = append(same, x)
		}
	}
	return same
}

func Differs(a, b []string) []string {
	mb := make(map[string]bool, len(b))
	for _, x := range b {
		mb[x] = true
	}
	var differ []string
	for _, x := range a {
		if _, found := mb[x]; !found {
			differ = append(differ, x)
		}
	}
	return differ
}

func String(v string) *string {
	return &v
}

func IsEmpty(v string) bool {
	return v == "" || v == "undefined"
}

func IsValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func TruncateText(str string, length int) string {
	if length <= 0 {
		return ""
	}

	truncated := ""
	count := 0
	for _, char := range str {
		truncated += string(char)
		count++
		if count >= length {
			break
		}
	}
	return truncated
}

func Singularize(s string) string {
	return pluralize.NewClient().Singular(s)
}

func Pluralize(s string) string {
	return pluralize.NewClient().Plural(s)
}

func UpperSinglarize(s string) string {
	return strings.Title(Singularize(s))
}

func UpperPluralize(s string) string {
	return strings.Title(Pluralize(s))
}

func LowerSinglarize(s string) string {
	return strings.ToLower(Singularize(s))
}

func LowerPluralize(s string) string {
	return strings.ToLower(Pluralize(s))
}

func Avatar(s string) string {
	return fmt.Sprintf("https://avatars.dicebear.com/api/pixel-art/%s.svg", s)
}

func ParseGraphResult(result *rg.QueryResult) []interface{} {
	itemIds := make([]interface{}, 0)
	if result != nil {
		for result.Next() { // Next returns true until the iterator is depleted.
			// Get the current Record.
			r := result.Record()
			// Entries in the Record can be accessed by index or key.
			record := ConvertInterfaceToMap(ConvertInterfaceToMap(r.GetByIndex(0))["Properties"])
			//if record["id"] != "--" { //TODO: hacky-none-fix
			itemIds = append(itemIds, record["id"])
			//}

		}
	}
	return itemIds
}

func ParseGraphResultWithStrIDs(result *rg.QueryResult) []string {
	itemIds := make([]string, 0)
	if result != nil {
		for result.Next() {
			r := result.Record()
			record := ConvertInterfaceToMap(ConvertInterfaceToMap(r.GetByIndex(0))["Properties"])
			itemIds = append(itemIds, record["id"].(string))
		}
	}
	return itemIds
}

func ExpvarGet(key string) string {
	val := expvar.Get(key)
	if val != nil {
		return strings.Trim(val.String(), "\"")
	}
	return ""
}

func NotEmpty(param string) bool {
	return param != "00000000-0000-0000-0000-000000000000" && param != "" && param != "undefined" && param != "nil"
}

func GenieID(entityID, itemID string) *string {
	var genieID *string
	if entityID != "" && itemID != "" {
		genieIDStr := fmt.Sprintf("%s#%s", entityID, itemID)
		genieID = &genieIDStr
	}
	return genieID
}

func PickGenieID(source map[string][]string) *string {
	for k, v := range source {
		if len(v) > 0 { // sending the first item. Because only one creator must be the source when the automation runs
			return GenieID(k, v[0])
		}
	}
	return nil
}

func AccountAsHost(accName string) string {
	accName = strings.ReplaceAll(accName, " ", "_")
	return strings.ToLower(accName)
}
