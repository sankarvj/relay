package util

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
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
		log.Println("expected error occured when marshalling on `ConvertInterfaceToMap` sending empty map. error:", err)
		return itemMap
	}
	err = json.Unmarshal(jsonbody, &itemMap)
	if err != nil {
		log.Println("expected error occured when unmarshalling on `ConvertInterfaceToMap` sending empty map. error:", err)
	}
	return itemMap
}

func ConvertStrToInt(s string) int {
	if i, err := strconv.Atoi(s); err == nil {
		return i
	}
	log.Printf("expected error occured when converting string %s to int on `ConvertStrToInt` sending zero\n", s)
	return 0
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
	return email[:at]
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
