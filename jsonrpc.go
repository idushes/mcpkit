package main

import (
	"encoding/json"
)

// Version is the JSON-RPC protocol version
const Version = "2.0"

// Request represents a JSON-RPC request object
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
	ID      interface{}     `json:"id,omitempty"`
}

// NewRequest creates a new Request with the specified method, parameters, and ID
func NewRequest(method string, params interface{}, id interface{}) (*Request, error) {
	if method == "" {
		return nil, StdError(ErrInvalidRequest)
	}

	var paramsJSON json.RawMessage
	if params != nil {
		data, err := json.Marshal(params)
		if err != nil {
			return nil, err
		}
		paramsJSON = json.RawMessage(data)
	}

	return &Request{
		JSONRPC: Version,
		Method:  method,
		Params:  paramsJSON,
		ID:      id,
	}, nil
}

// NewNotification creates a new notification Request with the specified method and parameters
func NewNotification(method string, params interface{}) (*Request, error) {
	return NewRequest(method, params, nil)
}

// IsNotification returns true if the request is a notification (has no ID)
func (r *Request) IsNotification() bool {
	return r.ID == nil
}

// Response represents a JSON-RPC response object
type Response struct {
	JSONRPC string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *Error          `json:"error,omitempty"`
	ID      interface{}     `json:"id"`
}

// NewResponse creates a new Response with the specified result and ID
func NewResponse(result interface{}, id interface{}) (*Response, error) {
	var resultJSON json.RawMessage
	if result != nil {
		data, err := json.Marshal(result)
		if err != nil {
			return nil, err
		}
		resultJSON = json.RawMessage(data)
	}

	return &Response{
		JSONRPC: Version,
		Result:  resultJSON,
		ID:      id,
	}, nil
}

// NewErrorResponse creates a new Response with the specified error and ID
func NewErrorResponse(err *Error, id interface{}) *Response {
	return &Response{
		JSONRPC: Version,
		Error:   err,
		ID:      id,
	}
}

// Error represents a JSON-RPC error object
type Error struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// Error returns a string representation of the error
func (e *Error) Error() string {
	return e.Message
}

// StdError returns a new Error with standard error message
func StdError(code int) *Error {
	return &Error{
		Code:    code,
		Message: ErrorMessage(code),
	}
}

// NewError creates a custom Error with specified code, message, and optional data
func NewError(code int, message string, data interface{}) (*Error, error) {
	var dataJSON json.RawMessage
	if data != nil {
		bytes, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		dataJSON = json.RawMessage(bytes)
	}

	return &Error{
		Code:    code,
		Message: message,
		Data:    dataJSON,
	}, nil
}

// Standard error codes as defined by JSON-RPC 2.0 spec
const (
	// Standard JSON-RPC 2.0 errors
	ErrParse          = -32700
	ErrInvalidRequest = -32600
	ErrMethodNotFound = -32601
	ErrInvalidParams  = -32602
	ErrInternal       = -32603
	// Server error range
	ErrServerErrorStart = -32099
	ErrServerErrorEnd   = -32000
)

// ErrorMessage returns the standard message for a given error code
func ErrorMessage(code int) string {
	switch code {
	case ErrParse:
		return "Parse error"
	case ErrInvalidRequest:
		return "Invalid Request"
	case ErrMethodNotFound:
		return "Method not found"
	case ErrInvalidParams:
		return "Invalid params"
	case ErrInternal:
		return "Internal error"
	default:
		if code >= ErrServerErrorStart && code <= ErrServerErrorEnd {
			return "Server error"
		}
		return "Unknown error"
	}
}

// BatchRequest represents a batch of JSON-RPC requests
type BatchRequest []*Request

// BatchResponse represents a batch of JSON-RPC responses
type BatchResponse []*Response
