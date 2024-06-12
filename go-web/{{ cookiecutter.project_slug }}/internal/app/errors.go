package app

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
)

type MyError struct {
	errs     validator.ValidationErrors
	httpCode int
	Code     int    `json:"code"`
	Message  string `json:"message"`
}

func NewMyError(err error, code int) *MyError {
	return NewMyErrorWithHTTPCode(err, code, http.StatusBadRequest)
}

func NewMyErrorWithHTTPCode(err error, code int, httpCode int) *MyError {
	verr, ok := err.(validator.ValidationErrors)
	if !ok {
		verr = nil
	}

	msg := fmt.Sprintf("%s: %s", GetMessageFromErrorCodeMap(code), err.Error())
	return &MyError{errs: verr, Code: code, Message: msg, httpCode: httpCode}
}

func (m *MyError) Error() string {
	return fmt.Sprintf("%d: %s", m.Code, m.Message)
}

func RenderError(w http.ResponseWriter, err error) {
	var merr *MyError
	var ok bool

	merr, ok = err.(*MyError)
	if !ok {
		merr = NewMyError(err, ErrorCodeUnknown)
	}

	w.WriteHeader(merr.httpCode)
	json.NewEncoder(w).Encode(merr)
}

func GetMessageFromErrorCodeMap(code int) string {
	if msg, ok := errorCodeMessageMap[code]; ok {
		return msg
	}
	return ""
}

const (
	ErrorCodeInvalidPassword int = 1000
	ErrorCodeUnknown         int = 9999
)

// errorCodeMessageMap is a error code map manager
var errorCodeMessageMap = map[int]string{
	// 1000 - 2000 for user relevant error codes
	ErrorCodeInvalidPassword: "invalid password",

	// 2000 - 3000 for xxx

	// database error
	ErrorCodeUnknown: "unknown error",
}
