package main

import (
	"encoding/json"
)

// MCP-specific types for the Model Context Protocol

// MCPAction defines the standard actions for MCP
type MCPAction string

// Standard MCP actions
const (
	MCPActionSubmit  MCPAction = "submit"
	MCPActionStream  MCPAction = "stream"
	MCPActionExecute MCPAction = "execute"
	MCPActionCancel  MCPAction = "cancel"
)

// MCPStatus defines the status codes for MCP responses
type MCPStatus string

// Standard MCP status values
const (
	MCPStatusSuccess MCPStatus = "success"
	MCPStatusError   MCPStatus = "error"
	MCPStatusPartial MCPStatus = "partial"
)

// MCPRequest extends the JSON-RPC Request with MCP-specific fields
type MCPRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Action  MCPAction       `json:"action"`
	Params  json.RawMessage `json:"params,omitempty"`
	Context interface{}     `json:"context,omitempty"`
	Tool    string          `json:"tool,omitempty"`
	ID      interface{}     `json:"id,omitempty"`
}

// MCPResponse extends the JSON-RPC Response with MCP-specific fields
type MCPResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	Status  MCPStatus       `json:"status"`
	Data    json.RawMessage `json:"data,omitempty"`
	Error   *Error          `json:"error,omitempty"`
	Context interface{}     `json:"context,omitempty"`
	ID      interface{}     `json:"id"`
}

// NewMCPRequest creates a new MCPRequest with the specified parameters
func NewMCPRequest(action MCPAction, params interface{}, context interface{}, tool string, id interface{}) (*MCPRequest, error) {
	var paramsJSON json.RawMessage
	if params != nil {
		data, err := json.Marshal(params)
		if err != nil {
			return nil, err
		}
		paramsJSON = json.RawMessage(data)
	}

	return &MCPRequest{
		JSONRPC: Version,
		Method:  string(action), // For backward compatibility
		Action:  action,
		Params:  paramsJSON,
		Context: context,
		Tool:    tool,
		ID:      id,
	}, nil
}

// NewMCPResponse creates a new MCPResponse with the specified parameters
func NewMCPResponse(status MCPStatus, data interface{}, context interface{}, id interface{}) (*MCPResponse, error) {
	var dataJSON json.RawMessage
	if data != nil {
		bytes, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		dataJSON = json.RawMessage(bytes)
	}

	return &MCPResponse{
		JSONRPC: Version,
		Status:  status,
		Data:    dataJSON,
		Context: context,
		ID:      id,
	}, nil
}

// NewMCPErrorResponse creates a new MCPResponse with an error
func NewMCPErrorResponse(err *Error, context interface{}, id interface{}) *MCPResponse {
	return &MCPResponse{
		JSONRPC: Version,
		Status:  MCPStatusError,
		Error:   err,
		Context: context,
		ID:      id,
	}
}

// MCP-specific error codes
const (
	ErrMCPActionNotSupported = -33001
	ErrMCPToolNotAvailable   = -33002
	ErrMCPContextInvalid     = -33003
	ErrMCPExecutionFailed    = -33004
)

// MCPErrorMessage returns the MCP-specific error message for a given error code
func MCPErrorMessage(code int) string {
	switch code {
	case ErrMCPActionNotSupported:
		return "Action not supported"
	case ErrMCPToolNotAvailable:
		return "Tool not available"
	case ErrMCPContextInvalid:
		return "Context invalid"
	case ErrMCPExecutionFailed:
		return "Execution failed"
	default:
		return ErrorMessage(code)
	}
}
