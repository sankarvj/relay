package handlers

import (
	"fmt"
	"log"

	"gitlab.com/vjsideprojects/relay/internal/platform/util"
)

type ErrorResponse struct {
	Type     int                     `json:"type"`
	ErrorMap map[string]ErrorPayload `json:"error_map"`
	Message  string                  `json:"message"`
}

type ErrorPayload struct {
	Message  string   `json:"message"`
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
		Message:  "This field cannot left blank",
		Solution: "Please choose a valid input",
	}
}

func uniqueErrorPayload(value interface{}) ErrorPayload {
	return ErrorPayload{
		Message:  "Duplicate item with same value exists",
		Solution: fmt.Sprintf("Please choose a different value for %s", value),
		Value:    value.(string),
	}
}

func uniqueErrorPayloads(value []interface{}) ErrorPayload {
	return ErrorPayload{
		Message:  "Duplicate item with same value exists",
		Solution: fmt.Sprintf("Please choose a different value for %s", value),
		ValueArr: util.ConvertSliceTypeRev(value),
	}
}

//Errors

func requiredError(errorMap map[string]ErrorPayload) *ErrorResponse {
	return &ErrorResponse{
		Type:     RequiredValidationError,
		ErrorMap: errorMap,
		Message:  "reqired field validation failed",
	}
}

func validationError(errorMap map[string]ErrorPayload) *ErrorResponse {
	return &ErrorResponse{
		Type:     UniqueValidationError,
		ErrorMap: errorMap,
		Message:  "unique validation failed",
	}
}

func unexpectedError(err error) *ErrorResponse {
	log.Println("unexpected error occurred ", err)
	return &ErrorResponse{
		Type:    UnexpectedError,
		Message: fmt.Sprintf("unexpected error occured:  %s", err.Error()),
	}
}
