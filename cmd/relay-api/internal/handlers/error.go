package handlers

import (
	"fmt"
	"log"

	"gitlab.com/vjsideprojects/relay/internal/platform/util"
)

type ErrorResponse struct {
	Type     int                     `json:"type"`
	ErrorMap map[string]ErrorPayload `json:"error_map"`
	Error    string                  `json:"error"`
}

type ErrorPayload struct {
	Error    string   `json:"message"`
	Solution string   `json:"solution"`
	Value    string   `json:"value"`
	ValueArr []string `json:"value_arr"`
}

const (
	UnexpectedError         int = 100
	UniqueValidationError   int = 1000
	RequiredValidationError int = 1001
)

//Error payloads

func requiredErrorPayload() ErrorPayload {
	return ErrorPayload{
		Error:    "This field cannot left blank",
		Solution: "Please choose a valid input",
	}
}

func uniqueErrorPayload(value interface{}) ErrorPayload {
	return ErrorPayload{
		Error:    "Duplicate item with same value exists",
		Solution: fmt.Sprintf("Please choose a different value for %s", value),
		Value:    value.(string),
	}
}

func uniqueErrorPayloads(value []interface{}) ErrorPayload {
	return ErrorPayload{
		Error:    "Duplicate item with same value exists",
		Solution: fmt.Sprintf("Please choose a different value for %s", value),
		ValueArr: util.ConvertSliceTypeRev(value),
	}
}

//Errors

func requiredError(errorMap map[string]ErrorPayload) *ErrorResponse {
	return &ErrorResponse{
		Type:     RequiredValidationError,
		ErrorMap: errorMap,
		Error:    "Reqired field validation failed",
	}
}

func validationError(errorMap map[string]ErrorPayload) *ErrorResponse {
	return &ErrorResponse{
		Type:     UniqueValidationError,
		ErrorMap: errorMap,
		Error:    "Unique validation failed",
	}
}

func uniqueErrorIsInvalid(erroredIds []string, id string) bool {
	var invalid bool
	for _, erroredID := range erroredIds {
		if erroredID == id {
			invalid = true
		}
	}
	return invalid
}

func unexpectedError(err error) *ErrorResponse {
	log.Println("***> unexpected error occurred ", err)
	return &ErrorResponse{
		Type:  UnexpectedError,
		Error: fmt.Sprintf("unexpected error occured:  %s", err.Error()),
	}
}
