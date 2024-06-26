// Package schemas provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/oapi-codegen/oapi-codegen/v2 version v2.3.0 DO NOT EDIT.
package schemas

// BasicError The basic structure for error response
type BasicError struct {
	// Code The http status code
	Code int `json:"code"`

	// Message The error message indicating what the issue is
	Message string `json:"message"`
}

// CreateUserRequest defines model for CreateUserRequest.
type CreateUserRequest struct {
	// Email email address
	Email string `json:"email" validate:"required,email,max=256"`

	// Name user display name
	Name string `json:"name" validate:"required,max=24"`

	// Password password
	Password string `json:"password"`

	// RepeatedPassword repeated password
	RepeatedPassword string `json:"repeated_password"`
}

// PostUsersJSONRequestBody defines body for PostUsers for application/json ContentType.
type PostUsersJSONRequestBody = CreateUserRequest
