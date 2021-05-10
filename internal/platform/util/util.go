package util

import (
	"encoding/json"
	"fmt"
	"log"
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
		log.Println("Error occured duting the interface marshal", err)
		return itemMap
	}
	err = json.Unmarshal(jsonbody, &itemMap)
	if err != nil {
		log.Println("Error occured duting the interface unmarshal", err)
	}
	return itemMap
}

func Contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
