package util

import (
	"encoding/json"
	"fmt"
	"log"
	"net/mail"
	"strconv"
	"strings"
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

func SubDomainInEmail(email string) string {
	at := strings.LastIndex(email, "@")
	if at >= 0 {
		username, domain := email[:at], email[at+1:]
		fmt.Printf("Username: %s, Domain: %s\n", username, domain)
		components := strings.Split(domain, ".")
		return components[0]
	} else {
		fmt.Printf("Error: %s is an invalid email address\n", email)
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
