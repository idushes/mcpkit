package main

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestNewRequest(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		params     interface{}
		id         interface{}
		wantErr    bool
		wantMethod string
		wantID     interface{}
	}{
		{
			name:       "basic request",
			method:     "test_method",
			params:     map[string]string{"key": "value"},
			id:         1,
			wantErr:    false,
			wantMethod: "test_method",
			wantID:     1,
		},
		{
			name:       "request with string ID",
			method:     "test_method",
			params:     []int{1, 2, 3},
			id:         "request-id",
			wantErr:    false,
			wantMethod: "test_method",
			wantID:     "request-id",
		},
		{
			name:       "request with nil params",
			method:     "test_method",
			params:     nil,
			id:         1,
			wantErr:    false,
			wantMethod: "test_method",
			wantID:     1,
		},
		{
			name:       "request with empty method",
			method:     "",
			params:     nil,
			id:         1,
			wantErr:    true,
			wantMethod: "",
			wantID:     1,
		},
		{
			name:       "request with special characters in method",
			method:     "!@#$%^&*()",
			params:     nil,
			id:         1,
			wantErr:    false,
			wantMethod: "!@#$%^&*()",
			wantID:     1,
		},
		{
			name:       "request with large numeric ID",
			method:     "test_method",
			params:     nil,
			id:         9223372036854775807,
			wantErr:    false,
			wantMethod: "test_method",
			wantID:     9223372036854775807,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := NewRequest(tt.method, tt.params, tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			if req.JSONRPC != Version {
				t.Errorf("req.JSONRPC = %v, want %v", req.JSONRPC, Version)
			}
			if req.Method != tt.wantMethod {
				t.Errorf("req.Method = %v, want %v", req.Method, tt.wantMethod)
			}
			if !reflect.DeepEqual(req.ID, tt.wantID) {
				t.Errorf("req.ID = %v, want %v", req.ID, tt.wantID)
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

func TestNewNotification(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		params     interface{}
		wantErr    bool
		wantMethod string
	}{
		{
			name:       "basic notification",
			method:     "update",
			params:     map[string]string{"status": "ok"},
			wantErr:    false,
			wantMethod: "update",
		},
		{
			name:       "notification with array params",
			method:     "notify",
			params:     []string{"a", "b", "c"},
			wantErr:    false,
			wantMethod: "notify",
		},
		{
			name:       "notification with nil params",
			method:     "ping",
			params:     nil,
			wantErr:    false,
			wantMethod: "ping",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := NewNotification(tt.method, tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewNotification() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			if req.JSONRPC != Version {
				t.Errorf("req.JSONRPC = %v, want %v", req.JSONRPC, Version)
			}
			if req.Method != tt.wantMethod {
				t.Errorf("req.Method = %v, want %v", req.Method, tt.wantMethod)
			}
			if req.ID != nil {
				t.Errorf("req.ID = %v, want nil", req.ID)
			}

			if !req.IsNotification() {
				t.Error("IsNotification() returned false for a notification")
			}
		})
	}
}

func TestIsNotification(t *testing.T) {
	tests := []struct {
		name string
		req  *Request
		want bool
	}{
		{
			name: "notification request",
			req:  &Request{JSONRPC: Version, Method: "update", ID: nil},
			want: true,
		},
		{
			name: "standard request with ID",
			req:  &Request{JSONRPC: Version, Method: "get", ID: 1},
			want: false,
		},
		{
			name: "request with string ID",
			req:  &Request{JSONRPC: Version, Method: "get", ID: "abc"},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.req.IsNotification(); got != tt.want {
				t.Errorf("Request.IsNotification() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewResponse(t *testing.T) {
	tests := []struct {
		name    string
		result  interface{}
		id      interface{}
		wantErr bool
	}{
		{
			name:    "response with string result",
			result:  "success",
			id:      1,
			wantErr: false,
		},
		{
			name:    "response with object result",
			result:  map[string]interface{}{"name": "value", "count": 42},
			id:      "request-id",
			wantErr: false,
		},
		{
			name:    "response with nil result",
			result:  nil,
			id:      1,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := NewResponse(tt.result, tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			if resp.JSONRPC != Version {
				t.Errorf("resp.JSONRPC = %v, want %v", resp.JSONRPC, Version)
			}
			if !reflect.DeepEqual(resp.ID, tt.id) {
				t.Errorf("resp.ID = %v, want %v", resp.ID, tt.id)
			}
			if resp.Error != nil {
				t.Errorf("resp.Error = %v, want nil", resp.Error)
			}

			if tt.result != nil {
				var result interface{}
				if err := json.Unmarshal(resp.Result, &result); err != nil {
					t.Errorf("Failed to unmarshal result: %v", err)
				}
			}
		})
	}
}

func TestNewErrorResponse(t *testing.T) {
	tests := []struct {
		name      string
		err       *Error
		id        interface{}
		wantCode  int
		wantError string
	}{
		{
			name:      "standard error response",
			err:       StdError(ErrInvalidRequest),
			id:        1,
			wantCode:  ErrInvalidRequest,
			wantError: "Invalid Request",
		},
		{
			name:      "custom error response",
			err:       &Error{Code: -1, Message: "Custom error"},
			id:        "abc",
			wantCode:  -1,
			wantError: "Custom error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := NewErrorResponse(tt.err, tt.id)

			if resp.JSONRPC != Version {
				t.Errorf("resp.JSONRPC = %v, want %v", resp.JSONRPC, Version)
			}
			if !reflect.DeepEqual(resp.ID, tt.id) {
				t.Errorf("resp.ID = %v, want %v", resp.ID, tt.id)
			}
			if resp.Result != nil {
				t.Errorf("resp.Result = %v, want nil", resp.Result)
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

func TestStdError(t *testing.T) {
	tests := []struct {
		name      string
		code      int
		wantCode  int
		wantError string
	}{
		{
			name:      "parse error",
			code:      ErrParse,
			wantCode:  ErrParse,
			wantError: "Parse error",
		},
		{
			name:      "invalid request",
			code:      ErrInvalidRequest,
			wantCode:  ErrInvalidRequest,
			wantError: "Invalid Request",
		},
		{
			name:      "method not found",
			code:      ErrMethodNotFound,
			wantCode:  ErrMethodNotFound,
			wantError: "Method not found",
		},
		{
			name:      "invalid params",
			code:      ErrInvalidParams,
			wantCode:  ErrInvalidParams,
			wantError: "Invalid params",
		},
		{
			name:      "internal error",
			code:      ErrInternal,
			wantCode:  ErrInternal,
			wantError: "Internal error",
		},
		{
			name:      "server error",
			code:      ErrServerErrorStart,
			wantCode:  ErrServerErrorStart,
			wantError: "Server error",
		},
		{
			name:      "unknown error",
			code:      -1,
			wantCode:  -1,
			wantError: "Unknown error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := StdError(tt.code)

			if err.Code != tt.wantCode {
				t.Errorf("err.Code = %v, want %v", err.Code, tt.wantCode)
			}
			if err.Message != tt.wantError {
				t.Errorf("err.Message = %v, want %v", err.Message, tt.wantError)
			}
			if err.Data != nil {
				t.Errorf("err.Data = %v, want nil", err.Data)
			}
		})
	}
}

func TestNewError(t *testing.T) {
	tests := []struct {
		name        string
		code        int
		message     string
		data        interface{}
		wantErr     bool
		wantCode    int
		wantMessage string
	}{
		{
			name:        "custom error without data",
			code:        -1000,
			message:     "Custom error message",
			data:        nil,
			wantErr:     false,
			wantCode:    -1000,
			wantMessage: "Custom error message",
		},
		{
			name:        "custom error with data",
			code:        -1001,
			message:     "Error with data",
			data:        map[string]string{"detail": "Additional information"},
			wantErr:     false,
			wantCode:    -1001,
			wantMessage: "Error with data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err, createErr := NewError(tt.code, tt.message, tt.data)
			if (createErr != nil) != tt.wantErr {
				t.Errorf("NewError() error = %v, wantErr %v", createErr, tt.wantErr)
				return
			}
			if createErr != nil {
				return
			}

			if err.Code != tt.wantCode {
				t.Errorf("err.Code = %v, want %v", err.Code, tt.wantCode)
			}
			if err.Message != tt.wantMessage {
				t.Errorf("err.Message = %v, want %v", err.Message, tt.wantMessage)
			}

			if tt.data != nil {
				if err.Data == nil {
					t.Error("err.Data is nil, want non-nil")
				} else {
					var data interface{}
					if unmarshalErr := json.Unmarshal(err.Data, &data); unmarshalErr != nil {
						t.Errorf("Failed to unmarshal data: %v", unmarshalErr)
					}
				}
			}
		})
	}
}

func TestErrorMessage(t *testing.T) {
	tests := []struct {
		name string
		code int
		want string
	}{
		{
			name: "parse error",
			code: ErrParse,
			want: "Parse error",
		},
		{
			name: "invalid request",
			code: ErrInvalidRequest,
			want: "Invalid Request",
		},
		{
			name: "method not found",
			code: ErrMethodNotFound,
			want: "Method not found",
		},
		{
			name: "invalid params",
			code: ErrInvalidParams,
			want: "Invalid params",
		},
		{
			name: "internal error",
			code: ErrInternal,
			want: "Internal error",
		},
		{
			name: "server error - start range",
			code: ErrServerErrorStart,
			want: "Server error",
		},
		{
			name: "server error - middle of range",
			code: (ErrServerErrorStart + ErrServerErrorEnd) / 2,
			want: "Server error",
		},
		{
			name: "server error - end range",
			code: ErrServerErrorEnd,
			want: "Server error",
		},
		{
			name: "unknown error",
			code: -1,
			want: "Unknown error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ErrorMessage(tt.code); got != tt.want {
				t.Errorf("ErrorMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJSONMarshalRequest(t *testing.T) {
	req, _ := NewRequest("test_method", map[string]string{"key": "value"}, 1)

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	var unmarshaledReq Request
	if err := json.Unmarshal(data, &unmarshaledReq); err != nil {
		t.Fatalf("Failed to unmarshal request: %v", err)
	}

	if unmarshaledReq.JSONRPC != req.JSONRPC {
		t.Errorf("JSONRPC = %v, want %v", unmarshaledReq.JSONRPC, req.JSONRPC)
	}
	if unmarshaledReq.Method != req.Method {
		t.Errorf("Method = %v, want %v", unmarshaledReq.Method, req.Method)
	}

	// For numeric IDs, JSON unmarshal converts to float64
	numID, isNum := req.ID.(int)
	if isNum {
		floatID, ok := unmarshaledReq.ID.(float64)
		if !ok {
			t.Errorf("Unmarshaled ID = %T, want float64", unmarshaledReq.ID)
		} else if float64(numID) != floatID {
			t.Errorf("ID = %v, want %v", floatID, numID)
		}
	} else if !reflect.DeepEqual(unmarshaledReq.ID, req.ID) {
		t.Errorf("ID = %v, want %v", unmarshaledReq.ID, req.ID)
	}
}

func TestJSONMarshalResponse(t *testing.T) {
	resp, _ := NewResponse("result", 1)

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal response: %v", err)
	}

	var unmarshaledResp Response
	if err := json.Unmarshal(data, &unmarshaledResp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if unmarshaledResp.JSONRPC != resp.JSONRPC {
		t.Errorf("JSONRPC = %v, want %v", unmarshaledResp.JSONRPC, resp.JSONRPC)
	}

	// For numeric IDs, JSON unmarshal converts to float64
	numID, isNum := resp.ID.(int)
	if isNum {
		floatID, ok := unmarshaledResp.ID.(float64)
		if !ok {
			t.Errorf("Unmarshaled ID = %T, want float64", unmarshaledResp.ID)
		} else if float64(numID) != floatID {
			t.Errorf("ID = %v, want %v", floatID, numID)
		}
	} else if !reflect.DeepEqual(unmarshaledResp.ID, resp.ID) {
		t.Errorf("ID = %v, want %v", unmarshaledResp.ID, resp.ID)
	}
}
