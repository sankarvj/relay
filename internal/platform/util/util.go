package util

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
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
