package apierror

import (
	"encoding/json"
	"fmt"
)

type ApiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *ApiError) Error() string {
	errByteArr, err := json.Marshal(e)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%s", errByteArr)
}

func TestError() error {
	e := &ApiError{
		Code:    "XXX",
		Message: "Test Error"}
	return e
}

func NotFoundError() error {
	e := &ApiError{
		Code:    "1000",
		Message: "Endpoint Not Found"}
	return e
}

func NotImplementedError() error {
	e := &ApiError{
		Code:    "1001",
		Message: "Api endpoint exists but is not yet implemented"}
	return e
}

func InvalidIDError() error {
	e := &ApiError{
		Code:    "1002",
		Message: "Invalid object ID"}
	return e
}

func InvalidRequestPayload() error {
	e := &ApiError{
		Code:    "1003",
		Message: "Invalid Request payload"}
	return e
}

func QueryParamMissing() error {
	e := &ApiError{
		Code:    "1004",
		Message: "Query param specified but value missing"}
	return e
}

func ParamStartBadValue() error {
	e := &ApiError{
		Code:    "1005",
		Message: "Query param start must be positive integer"}
	return e
}

func ParamCountBadValue() error {
	e := &ApiError{
		Code:    "1006",
		Message: "Query param count must be positive integer"}
	return e
}

func MissingRequiredFieldsPayload() error {
	e := &ApiError{
		Code:    "1007",
		Message: "Missing required fields"}
	return e
}

func InvalidBasicAuthHeaders() error {
	e := &ApiError{
		Code:    "1008",
		Message: "Invalid basic auth headers"}
	return e
}

func BadClientSecret() error {
	e := &ApiError{
		Code:    "1009",
		Message: "Bad client secret"}
	return e
}

func InternalServerError() error {
	e := &ApiError{
		Code:    "1010",
		Message: "InternalServerError"}
	return e
}
