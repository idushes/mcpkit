package main

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestNewMCPRequest(t *testing.T) {
	tests := []struct {
		name       string
		action     MCPAction
		params     interface{}
		context    interface{}
		tool       string
		id         interface{}
		wantErr    bool
		wantAction MCPAction
		wantMethod string
		wantTool   string
		wantID     interface{}
	}{
		{
			name:       "submit action request",
			action:     MCPActionSubmit,
			params:     map[string]string{"key": "value"},
			context:    map[string]string{"ctx": "data"},
			tool:       "test-tool",
			id:         1,
			wantErr:    false,
			wantAction: MCPActionSubmit,
			wantMethod: "submit",
			wantTool:   "test-tool",
			wantID:     1,
		},
		{
			name:       "stream action request",
			action:     MCPActionStream,
			params:     []int{1, 2, 3},
			context:    nil,
			tool:       "",
			id:         "request-id",
			wantErr:    false,
			wantAction: MCPActionStream,
			wantMethod: "stream",
			wantTool:   "",
			wantID:     "request-id",
		},
		{
			name:       "execute action with nil params",
			action:     MCPActionExecute,
			params:     nil,
			context:    map[string]interface{}{"scope": "global"},
			tool:       "executor",
			id:         1,
			wantErr:    false,
			wantAction: MCPActionExecute,
			wantMethod: "execute",
			wantTool:   "executor",
			wantID:     1,
		},
		{
			name:       "cancel action with complex params",
			action:     MCPActionCancel,
			params:     map[string]interface{}{"id": 123, "reason": "user request"},
			context:    nil,
			tool:       "",
			id:         9223372036854775807,
			wantErr:    false,
			wantAction: MCPActionCancel,
			wantMethod: "cancel",
			wantTool:   "",
			wantID:     9223372036854775807,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := NewMCPRequest(tt.action, tt.params, tt.context, tt.tool, tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewMCPRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			if req.JSONRPC != Version {
				t.Errorf("req.JSONRPC = %v, want %v", req.JSONRPC, Version)
			}
			if req.Action != tt.wantAction {
				t.Errorf("req.Action = %v, want %v", req.Action, tt.wantAction)
			}
			if req.Method != tt.wantMethod {
				t.Errorf("req.Method = %v, want %v", req.Method, tt.wantMethod)
			}
			if req.Tool != tt.wantTool {
				t.Errorf("req.Tool = %v, want %v", req.Tool, tt.wantTool)
			}
			if !reflect.DeepEqual(req.ID, tt.wantID) {
				t.Errorf("req.ID = %v, want %v", req.ID, tt.wantID)
			}
			if !reflect.DeepEqual(req.Context, tt.context) {
				t.Errorf("req.Context = %v, want %v", req.Context, tt.context)
			}

			// Verify params marshaled correctly
			if tt.params != nil {
				var params interface{}
				if err := json.Unmarshal(req.Params, &params); err != nil {
					t.Errorf("Failed to unmarshal params: %v", err)
				}
			}
		})
	}
}

func TestNewMCPResponse(t *testing.T) {
	tests := []struct {
		name       string
		status     MCPStatus
		data       interface{}
		context    interface{}
		id         interface{}
		wantErr    bool
		wantStatus MCPStatus
		wantID     interface{}
	}{
		{
			name:       "success response",
			status:     MCPStatusSuccess,
			data:       "success data",
			context:    map[string]string{"ctx": "response-data"},
			id:         1,
			wantErr:    false,
			wantStatus: MCPStatusSuccess,
			wantID:     1,
		},
		{
			name:       "partial response",
			status:     MCPStatusPartial,
			data:       map[string]interface{}{"progress": 50, "message": "halfway there"},
			context:    nil,
			id:         "request-id",
			wantErr:    false,
			wantStatus: MCPStatusPartial,
			wantID:     "request-id",
		},
		{
			name:       "response with nil data",
			status:     MCPStatusSuccess,
			data:       nil,
			context:    map[string]int{"count": 42},
			id:         1,
			wantErr:    false,
			wantStatus: MCPStatusSuccess,
			wantID:     1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := NewMCPResponse(tt.status, tt.data, tt.context, tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewMCPResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			if resp.JSONRPC != Version {
				t.Errorf("resp.JSONRPC = %v, want %v", resp.JSONRPC, Version)
			}
			if resp.Status != tt.wantStatus {
				t.Errorf("resp.Status = %v, want %v", resp.Status, tt.wantStatus)
			}
			if !reflect.DeepEqual(resp.ID, tt.wantID) {
				t.Errorf("resp.ID = %v, want %v", resp.ID, tt.wantID)
			}
			if !reflect.DeepEqual(resp.Context, tt.context) {
				t.Errorf("resp.Context = %v, want %v", resp.Context, tt.context)
			}
			if resp.Error != nil {
				t.Errorf("resp.Error = %v, want nil", resp.Error)
			}

			if tt.data != nil {
				var data interface{}
				if err := json.Unmarshal(resp.Data, &data); err != nil {
					t.Errorf("Failed to unmarshal data: %v", err)
				}
			}
		})
	}
}

func TestNewMCPErrorResponse(t *testing.T) {
	tests := []struct {
		name       string
		err        *Error
		context    interface{}
		id         interface{}
		wantStatus MCPStatus
		wantCode   int
		wantError  string
	}{
		{
			name:       "standard error response",
			err:        StdError(ErrInvalidRequest),
			context:    nil,
			id:         1,
			wantStatus: MCPStatusError,
			wantCode:   ErrInvalidRequest,
			wantError:  "Invalid Request",
		},
		{
			name:       "custom error response",
			err:        &Error{Code: ErrMCPActionNotSupported, Message: "Action not supported"},
			context:    map[string]string{"source": "test"},
			id:         "abc",
			wantStatus: MCPStatusError,
			wantCode:   ErrMCPActionNotSupported,
			wantError:  "Action not supported",
		},
		{
			name:       "tool not available error",
			err:        &Error{Code: ErrMCPToolNotAvailable, Message: "Tool not available"},
			context:    nil,
			id:         42,
			wantStatus: MCPStatusError,
			wantCode:   ErrMCPToolNotAvailable,
			wantError:  "Tool not available",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := NewMCPErrorResponse(tt.err, tt.context, tt.id)

			if resp.JSONRPC != Version {
				t.Errorf("resp.JSONRPC = %v, want %v", resp.JSONRPC, Version)
			}
			if resp.Status != tt.wantStatus {
				t.Errorf("resp.Status = %v, want %v", resp.Status, tt.wantStatus)
			}
			if !reflect.DeepEqual(resp.ID, tt.id) {
				t.Errorf("resp.ID = %v, want %v", resp.ID, tt.id)
			}
			if !reflect.DeepEqual(resp.Context, tt.context) {
				t.Errorf("resp.Context = %v, want %v", resp.Context, tt.context)
			}
			if resp.Data != nil {
				t.Errorf("resp.Data = %v, want nil", resp.Data)
			}
			if resp.Error == nil {
				t.Error("resp.Error is nil, want non-nil")
				return
			}
			if resp.Error.Code != tt.wantCode {
				t.Errorf("resp.Error.Code = %v, want %v", resp.Error.Code, tt.wantCode)
			}
			if resp.Error.Message != tt.wantError {
				t.Errorf("resp.Error.Message = %v, want %v", resp.Error.Message, tt.wantError)
			}
		})
	}
}

func TestMCPErrorMessage(t *testing.T) {
	tests := []struct {
		name      string
		code      int
		wantError string
	}{
		{
			name:      "action not supported error",
			code:      ErrMCPActionNotSupported,
			wantError: "Action not supported",
		},
		{
			name:      "tool not available error",
			code:      ErrMCPToolNotAvailable,
			wantError: "Tool not available",
		},
		{
			name:      "context invalid error",
			code:      ErrMCPContextInvalid,
			wantError: "Context invalid",
		},
		{
			name:      "execution failed error",
			code:      ErrMCPExecutionFailed,
			wantError: "Execution failed",
		},
		{
			name:      "standard JSON-RPC error",
			code:      ErrInvalidRequest,
			wantError: "Invalid Request",
		},
		{
			name:      "unknown error",
			code:      -1,
			wantError: "Unknown error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := MCPErrorMessage(tt.code)
			if msg != tt.wantError {
				t.Errorf("MCPErrorMessage(%d) = %v, want %v", tt.code, msg, tt.wantError)
			}
		})
	}
}

func TestJSONMarshalMCPRequest(t *testing.T) {
	req, _ := NewMCPRequest(
		MCPActionExecute,
		map[string]string{"command": "test"},
		map[string]int{"priority": 1},
		"test-tool",
		42,
	)

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	// Unmarshal to verify JSON structure
	var unmarshaled map[string]interface{}
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Verify fields
	if unmarshaled["jsonrpc"] != Version {
		t.Errorf("jsonrpc = %v, want %v", unmarshaled["jsonrpc"], Version)
	}
	if unmarshaled["method"] != string(MCPActionExecute) {
		t.Errorf("method = %v, want %v", unmarshaled["method"], string(MCPActionExecute))
	}
	if unmarshaled["action"] != string(MCPActionExecute) {
		t.Errorf("action = %v, want %v", unmarshaled["action"], string(MCPActionExecute))
	}
	if unmarshaled["tool"] != "test-tool" {
		t.Errorf("tool = %v, want %v", unmarshaled["tool"], "test-tool")
	}
	if unmarshaled["id"].(float64) != 42 {
		t.Errorf("id = %v, want %v", unmarshaled["id"], 42)
	}

	// Check context
	context, ok := unmarshaled["context"].(map[string]interface{})
	if !ok {
		t.Fatal("context is not a map")
	}
	if priority, ok := context["priority"].(float64); !ok || priority != 1 {
		t.Errorf("context.priority = %v, want %v", context["priority"], 1)
	}

	// Check params
	params, ok := unmarshaled["params"].(map[string]interface{})
	if !ok {
		t.Fatal("params is not a map")
	}
	if command, ok := params["command"].(string); !ok || command != "test" {
		t.Errorf("params.command = %v, want %v", params["command"], "test")
	}
}

func TestJSONMarshalMCPResponse(t *testing.T) {
	resp, _ := NewMCPResponse(
		MCPStatusSuccess,
		map[string]string{"result": "completed"},
		map[string]bool{"final": true},
		42,
	)

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal response: %v", err)
	}

	// Unmarshal to verify JSON structure
	var unmarshaled map[string]interface{}
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Verify fields
	if unmarshaled["jsonrpc"] != Version {
		t.Errorf("jsonrpc = %v, want %v", unmarshaled["jsonrpc"], Version)
	}
	if unmarshaled["status"] != string(MCPStatusSuccess) {
		t.Errorf("status = %v, want %v", unmarshaled["status"], string(MCPStatusSuccess))
	}
	if unmarshaled["id"].(float64) != 42 {
		t.Errorf("id = %v, want %v", unmarshaled["id"], 42)
	}

	// Check context
	context, ok := unmarshaled["context"].(map[string]interface{})
	if !ok {
		t.Fatal("context is not a map")
	}
	if final, ok := context["final"].(bool); !ok || !final {
		t.Errorf("context.final = %v, want %v", context["final"], true)
	}

	// Check data
	respData, ok := unmarshaled["data"].(map[string]interface{})
	if !ok {
		t.Fatal("data is not a map")
	}
	if result, ok := respData["result"].(string); !ok || result != "completed" {
		t.Errorf("data.result = %v, want %v", respData["result"], "completed")
	}
}
