# Model Context Protocol (MCP) — Specification

## Table of Contents

- [Overview](#overview)
- [Goals](#goals)
- [Message Format](#message-format)
  - [MCP Request](#1-client--server-mcp-request)
  - [MCP Response](#2-server--client-mcp-response)
  - [Handling Large Payloads](#handling-large-payloads)
- [Transport Modes](#transport-modes)
  - [Stdio](#stdio)
  - [HTTP](#http)
  - [Server-Sent Events (SSE)](#server-sent-events-sse)
  - [WebSockets](#websockets)
- [Actions](#actions)
  - [Action Definition](#action-definition)
  - [Action Naming Conventions](#action-naming-conventions)
  - [Action Discovery](#action-discovery)
  - [Standard Actions](#standard-actions)
  - [Example Actions](#example-actions)
- [Examples](#examples)
  - [Example 1: Successful Tool Call](#example-1--successful-tool-call)
  - [Example 2: Error Response](#example-2--error-response)
  - [Example 3: Streaming Response](#example-3--streaming-response)
  - [Example 4: Batch Operations](#example-4--batch-operations)
- [Error Handling](#error-handling)
  - [Error Response Structure](#error-response-structure)
  - [Error Types (Recommendation)](#error-types-recommendation)
  - [Handling Partial Successes](#handling-partial-successes)
  - [Client-Side Error Handling](#client-side-error-handling)
- [Security Considerations](#security-considerations)
  - [Authentication](#authentication)
  - [Authorization](#authorization)
  - [Input Validation](#input-validation)
  - [Rate Limiting](#rate-limiting)
- [Versioning](#versioning)
  - [Semantic Versioning](#semantic-versioning)
  - [Backward Compatibility](#backward-compatibility)
  - [Protocol Evolution](#protocol-evolution)
- [Compliance Guidelines](#compliance-guidelines)
- [Implementation Guidance](#implementation-guidance)
  - [Language-Specific Recommendations](#language-specific-recommendations)
  - [Testing Strategies](#testing-strategies)
  - [Performance Optimization](#performance-optimization)
  - [Retry Strategies](#retry-strategies)
- [Summary](#summary)

---

## Overview

**Model Context Protocol (MCP)** is a lightweight, structured protocol designed to facilitate standardized communication between language models (LLMs), software agents, and external tools or services. It defines a simple JSON-based request/response format that can operate over various transport layers (e.g., stdio, HTTP, WebSockets, SSE). MCP promotes interoperability by decoupling components, allowing them to interact without requiring bespoke integrations.

---

## Goals

- **Standardization:** Provide a simple, consistent, and machine-readable format for interactions.
- **Interoperability:** Enable diverse tools, agents, and models to communicate effectively.
- **Flexibility:** Support invocation of actions with structured parameters.
- **Transport Agnosticism:** Function reliably across different communication channels like CLI, subprocesses, HTTP APIs, and streaming endpoints.
- **Extensibility:** Allow for optional fields and future protocol evolution.
- **Robustness:** Provide clear error handling patterns and recovery mechanisms.
- **Security:** Encourage secure implementation practices by design.

---

## Message Format

MCP messages **MUST** be encoded as JSON objects.

### 1. Client → Server (MCP Request)

An MCP Request must include the following fields:

- **`action`** (string, required): Clearly identifies the action to perform, following a consistent naming convention (e.g., `verb_noun`, `namespace.action`).
- **`params`** (object, required): Contains named parameters for the action. Even if no parameters are required, an empty object `{}` **MUST** be provided.
- **`tool`** (string, optional): Identifies the target tool or agent, useful in multi-tool environments.
- **`context`** (string, optional): Associates the request with a specific ongoing process or state.
- **`metadata`** (object, optional): Includes supplementary information such as protocol version, client identifiers, request IDs, or timestamps. Including a unique `request_id` is **RECOMMENDED** for debugging and correlation.

**Example MCP Request:**

```json
{
  "action": "file_system.read",
  "params": {
    "path": "/data/file.txt"
  },
  "tool": "filesystem_tool",
  "context": "session-123",
  "metadata": {
    "version": "1.0",
    "client_id": "agent-xyz",
    "request_id": "uuid-xyz-123",
    "timestamp": "2023-10-25T12:34:56Z"
  }
}
```

### 2. Server → Client (MCP Response)

A server response must clearly indicate success or failure:

- **`status`** (string, required): Either `"success"` or `"error"`.
- **`data`** (object, optional): Contains action results if successful.
- **`error`** (object, required if status is `"error"`): Provides detailed error information.
- **`metadata`** (object, optional): Echoes back the `request_id` and includes additional response details.

**Example Successful Response:**

```json
{
  "status": "success",
  "data": {
    "content": "File content here"
  },
  "metadata": {
    "server_id": "filesystem_tool",
    "execution_time_ms": 120,
    "request_id": "uuid-xyz-123"
  }
}
```

**Example Error Response:**

```json
{
  "status": "error",
  "error": {
    "type": "RESOURCE_NOT_FOUND",
    "message": "File '/data/file.txt' not found.",
    "details": { "path": "/data/file.txt" },
    "code": 404
  },
  "metadata": {
    "server_id": "filesystem_tool",
    "request_id": "uuid-xyz-123"
  }
}
```

### Handling Large Payloads

For large payloads, implement pagination, chunking, or external references. Clearly document maximum payload sizes and provide examples of paginated responses.

---

## Transport Modes

MCP is designed to be transport-agnostic. Common implementations include:

### Stdio

- **Communication:** The client writes the MCP Request JSON (typically followed by a newline `\n`) to the server process's `stdin`. The server writes the MCP Response JSON (typically followed by a newline `\n`) to its `stdout`.
- **Encoding:** UTF-8 encoding is **REQUIRED**.
- **Termination:** The server process should typically exit after sending its response, unless designed for persistent interaction. Clients should handle potential blocking reads and process termination.
- **Message Framing:** For binary data or multi-message exchanges, implementations **MAY** use length-prefixed framing (sending message length before each message) to avoid ambiguity.
- **Error Output:** Implementations **SHOULD** capture and structure stderr output as part of the response when possible, rather than letting it flow separately.

### HTTP

- **Endpoint:** Typically a dedicated endpoint, e.g., `POST /mcp`.
- **Request:** The client sends an HTTP POST request with the MCP Request JSON as the request body. The `Content-Type` header **MUST** be set to `application/json`.
- **Response:** The server replies with the MCP Response JSON as the response body. The `Content-Type` header **MUST** be `application/json`.
- **Status Codes:**
    - `200 OK`: Successful processing (even if `status` is `"error"` in the MCP response). The MCP response body indicates the actual outcome.
    - `400 Bad Request`: Malformed JSON request or invalid MCP structure (before action processing).
    - `401 Unauthorized`: Missing or invalid authentication credentials.
    - `403 Forbidden`: Authentication succeeded but the client lacks permission for the requested action.
    - `404 Not Found`: Endpoint not found or action not supported.
    - `405 Method Not Allowed`: If a method other than POST is used.
    - `413 Payload Too Large`: Request exceeds size limits.
    - `429 Too Many Requests`: Rate limit exceeded.
    - `500 Internal Server Error`: Server-side failure unrelated to the specific MCP action logic.
    - `503 Service Unavailable`: The MCP server is temporarily unable to handle requests.
- **Headers:** Implementations **SHOULD** support standard HTTP headers for features like caching, compression, and CORS.
- **Timeouts:** Clients and servers **SHOULD** implement appropriate connection and request timeouts, with timeout values documented.

### Server-Sent Events (SSE)

- **Use Case:** Suitable for streaming responses where results arrive incrementally (e.g., LLM token generation, long-running tasks providing progress updates).
- **Connection:** Client establishes an SSE connection to a specific endpoint. A request (potentially as query parameters or an initial message) triggers the stream.
- **Message Format:** Each message chunk is sent as an SSE `data` event.

SSE events for MCP **SHOULD** follow this structure:

```
event: message
data: {"type": "partial", "data": {"token": "The"}}

event: message
data: {"type": "partial", "data": {"token": " quick"}}

event: message
data: {"type": "partial", "data": {"token": " a"}}

event: message
data: {"type": "partial", "data": {"token": " time"}}

event: message
data: {"type": "complete", "status": "success", "data": {"full_text": "The quick brown fox"}}
```

or for errors:

```
event: message
data: {"type": "error", "error": {"type": "PROCESSING_ERROR", "message": "Failed to generate text"}}
```

- **End of Stream:** The server **MUST** signal the end of a complete logical MCP response with either:
  - A message with `"type": "complete"` containing the final result
  - A message with `"type": "error"` containing error details
  - Closing the SSE connection

- **Client Handling:** Clients **MUST** be prepared to:
  - Buffer and reassemble partial results
  - Handle connection drops (with potential reconnection strategy)
  - Process both incremental and complete message types

### WebSockets

- **Connection:** A persistent, bidirectional connection allows multiple request/response cycles without the overhead of establishing new connections.
- **Message Format:** MCP JSON messages are sent as WebSocket text frames. Binary frames **MAY** be used for binary data with appropriate documentation.
- **Multiple Actions:** A single WebSocket connection can support multiple ongoing MCP request/response interactions.
- **Correlation:** To match responses with requests when multiple actions are in flight, clients **SHOULD** include a unique `request_id` in each request's metadata, which servers **MUST** echo in the corresponding response.

Example WebSocket message flow:
```
→ {"action": "start_generation", "params": {"prompt": "Write a poem"}, "metadata": {"request_id": "req1"}}
← {"status": "success", "data": {"generation_id": "gen123"}, "metadata": {"request_id": "req1"}}
→ {"action": "get_generation_progress", "params": {"generation_id": "gen123"}, "metadata": {"request_id": "req2"}}
← {"status": "success", "data": {"progress": 0.5, "text_so_far": "Roses are red"}, "metadata": {"request_id": "req2"}}
```

- **Heartbeats:** For long-lived connections, implementations **SHOULD** support WebSocket ping/pong frames or application-level heartbeats.
- **Termination:** Either party may close the connection with appropriate WebSocket close codes. Clients **SHOULD** implement reconnection logic with exponential backoff.

---

## Actions

### Action Definition

Actions represent the core capabilities offered by an MCP server (tool or agent). Each action **MUST** have:

- **A unique name** (string): Identifies the action (e.g., `document_summarizer.summarize`).
- **A defined parameter schema:** Specifies the expected structure and types within the `params` object for requests invoking this action. JSON Schema is a recommended standard for defining this.
- **A defined output schema:** Specifies the expected structure and types within the `data` object for successful responses.
- **Documentation:** Human-readable description of the action's purpose, parameters, return values, and potential errors.

### Action Naming Conventions

Action names **SHOULD** follow these conventions:

1. **Namespacing:** Use dot notation to organize actions into logical groups (e.g., `file_system.read`, `file_system.write`).
2. **Verb-first naming:** Begin with a verb that describes the operation (e.g., `get_`, `create_`, `update_`, `delete_`, `search_`).
3. **Consistency:** Use consistent terminology across related actions.
4. **Specificity:** Names should clearly indicate the action's purpose without ambiguity.

Examples of well-formed action names:
- `document.summarize`
- `database.query`
- `search.web`
- `image.generate`
- `weather.get_forecast`

### Action Discovery

To facilitate interoperability, MCP servers *SHOULD* provide a mechanism for clients to discover available actions and their schemas. A common convention is to implement a built-in action:

- **Action:** `mcp.get_schema`
- **Params:** (Optional) `{"action_name": "specific_action"}` to get the schema for one action, or omit to get schemas for all actions.
- **Data (Success):** An object containing action schemas, potentially in JSON Schema format.

```json
// Request to discover schemas
{
  "action": "mcp.get_schema",
  "params": {}
}

// Response with schemas
{
  "status": "success",
  "data": {
    "actions": {
      "summarize_text": {
        "description": "Summarizes the input text.",
        "params_schema": { /* JSON Schema for params */ },
        "data_schema": { /* JSON Schema for data */ },
        "examples": [{
          "request": { "params": {"text": "Example text", "max_length": 100} },
          "response": { "data": {"summary": "Example summary"} }
        }]
      },
      "list_files": { /* ... */ }
    }
  }
}
```

### Standard Actions

Implementations **SHOULD** consider supporting these standard actions:

1. **Introspection Actions:**
   - `mcp.get_schema`: Returns schema information about available actions
   - `mcp.get_version`: Returns version information about the tool/server
   - `mcp.health_check`: Verifies the server is operational

2. **Common Patterns:**
   - CRUD operations should follow consistent naming: `[resource].create`, `[resource].read`, `[resource].update`, `[resource].delete`
   - List operations should support pagination: `[resource].list`
   - Search operations should have consistent parameter patterns: `[resource].search`

### Example Actions

- `search_documents`: Finds documents matching query terms.
- `database.query`: Executes a query against a database.
- `file_system.write`: Writes content to a file.
- `code_interpreter.execute`: Executes a snippet of code.
- `llm.generate`: Generates text from a prompt.
- `image.create`: Creates an image based on parameters.
- `translation.translate`: Translates text between languages.

---

## Examples

### Example 1 — Successful Tool Call

An LLM agent asks a file system tool to list files in a directory.

**Client Request:**

```json
{
  "action": "file_system.list_directory",
  "params": {
    "path": "/usr/data",
    "recursive": false
  },
  "context": "conversation-123",
  "metadata": {
    "client_id": "agent-alpha",
    "request_id": "req-456"
   }
}
```

**Server Response:**

```json
{
  "status": "success",
  "data": {
    "files": [
      {"name": "file1.txt", "type": "file", "size": 1024},
      {"name": "subdir", "type": "directory", "size": null}
    ],
    "path": "/usr/data"
  },
  "metadata": {
    "server_id": "fs-tool-prod",
    "execution_time_ms": 45,
    "request_id": "req-456"
  }
}
```

### Example 2 — Error Response

The client attempts to call an action with an invalid parameter.

**Client Request:**

```json
{
  "action": "calculator.add",
  "params": {
    "a": 5,
    "b": "three" // Invalid type
  },
  "metadata": {
    "request_id": "calc-123"
  }
}
```

**Server Response:**

```json
{
  "status": "error",
  "error": {
    "type": "VALIDATION_ERROR",
    "message": "Invalid type for parameter 'b'. Expected number, got string.",
    "details": {
      "parameter": "b",
      "expected_type": "number",
      "actual_type": "string"
    },
    "code": 400
  },
  "metadata": {
    "server_id": "calc-tool-v2",
    "request_id": "calc-123"
  }
}
```

### Example 3 — Streaming Response

Using SSE to stream generation results:

**Client Request (initial HTTP request to establish SSE):**
```
GET /mcp/stream?action=llm.generate&prompt=Write%20a%20story HTTP/1.1
Accept: text/event-stream
```

**Server Response (SSE stream):**
```
event: message
data: {"type": "partial", "data": {"token": "Once"}}

event: message
data: {"type": "partial", "data": {"token": " upon"}}

event: message
data: {"type": "partial", "data": {"token": " a"}}

event: message
data: {"type": "partial", "data": {"token": " time"}}

event: message
data: {"type": "complete", "status": "success", "data": {"full_text": "Once upon a time"}}
```

### Example 4 — Batch Operations

Request to perform multiple operations in a single call:

**Client Request:**
```json
{
  "action": "batch.execute",
  "params": {
    "operations": [
      {
        "action": "file_system.read",
        "params": { "path": "/data/file1.txt" }
      },
      {
        "action": "file_system.write",
        "params": { "path": "/data/file2.txt", "content": "New content" }
      }
    ],
    "continue_on_error": true
  },
  "metadata": {
    "request_id": "batch-xyz-789"
  }
}
```

**Server Response:**
```json
{
  "status": "success", 
  "data": {
    "results": [
      {
        "status": "success",
        "data": { "content": "File 1 content" }
      },
      {
        "status": "error",
        "error": {
          "type": "PERMISSION_DENIED",
          "message": "Write access denied to /data/file2.txt"
        }
      }
    ],
    "success_count": 1,
    "error_count": 1
  },
  "metadata": {
    "request_id": "batch-xyz-789"
  }
}
```

---

## Error Handling

Robust error handling is crucial for reliable MCP communication.

### Error Response Structure

When `status` is `"error"`, the `error` field **MUST** be present. It is **RECOMMENDED** that the `error` object contain:

- **`type`** (string, required): A machine-readable error code or category (e.g., `INVALID_PARAMETER`, `ACTION_NOT_FOUND`, `AUTHENTICATION_FAILED`, `INTERNAL_SERVER_ERROR`). Use uppercase snake case.
- **`message`** (string, required): A clear, human-readable description of the error.
- **`details`** (object, optional): An object containing additional context-specific information about the error (e.g., invalid parameter name, missing resource ID).
- **`code`** (number, optional): A numeric error code, often matching HTTP status codes for consistency.
- **`help_url`** (string, optional): A URL pointing to documentation about this error type and potential solutions.

### Error Types (Recommendation)

Standardizing error types enhances client-side handling. Consider categories like:

- `VALIDATION_ERROR`: Input data (e.g., `params`) failed validation.
- `ACTION_NOT_FOUND`: The requested `action` is not supported by the server.
- `PARAMETER_MISSING`: A required parameter was not provided.
- `AUTHENTICATION_FAILED`: Client credentials are missing or invalid.
- `AUTHORIZATION_FAILED`: Client is not permitted to perform the action.
- `RESOURCE_NOT_FOUND`: A required external resource (e.g., file, database entry) was not found.
- `RATE_LIMIT_EXCEEDED`: Client has made too many requests.
- `QUOTA_EXCEEDED`: Client has exceeded usage quota (e.g., tokens, storage).
- `INTERNAL_SERVER_ERROR`: An unexpected error occurred on the server.
- `TIMEOUT_ERROR`: The action execution exceeded the allowed time limit.
- `DEPENDENCY_ERROR`: A required external service is unavailable.
- `INVALID_STATE`: The action cannot be performed in the current state.
- `CONFLICT`: The requested operation conflicts with the current state.

### Handling Partial Successes

For operations that involve multiple items or steps, partial success scenarios should be handled consistently:

1. **Batch Operations:** 
   - Return an array of individual results with success/error status for each item
   - Include summary statistics (e.g., success_count, error_count)
   - Allow clients to control behavior with parameters like `continue_on_error`

2. **Transactional Operations:**
   - For operations that should be atomic, use proper error handling to roll back on failure
   - Clearly document which operations are transactional vs. best-effort

Example for partial success:
```json
{
  "status": "success",
  "data": {
    "results": [
      {"id": "item1", "status": "success"},
      {"id": "item2", "status": "error", "error": {"type": "VALIDATION_ERROR"}}
    ],
    "success_count": 1,
    "failure_count": 1
  }
}
```

### Client-Side Error Handling

Clients **SHOULD** implement the following error handling strategies:

1. **Validation**: Validate requests locally before sending when possible.
2. **Retries**: Implement exponential backoff for transient errors (e.g., rate limiting, temporary service unavailability).
3. **Fallbacks**: Consider alternative actions when primary actions fail.
4. **Timeouts**: Set appropriate timeouts and handle timeout errors gracefully.
5. **Logging**: Log error responses for debugging and monitoring.
6. **User Feedback**: Translate error responses into appropriate user-facing messages.

Example client retry logic:
```
function executeWithRetry(request, maxRetries = 3, baseDelay = 1000) {
  let attempt = 0;
  
  while (attempt < maxRetries) {
    try {
      const response = sendRequest(request);
      
      if (response.status === "success" || 
          (response.status === "error" && !isRetryableError(response.error))) {
        return response;
      }
      
      // Exponential backoff
      const delay = baseDelay * Math.pow(2, attempt);
      wait(delay);
      attempt++;
    } catch (e) {
      // Handle connection errors
      if (!isRetryableException(e) || attempt >= maxRetries - 1) {
        throw e;
      }
      
      const delay = baseDelay * Math.pow(2, attempt);
      wait(delay);
      attempt++;
    }
  }
}
```

---

## Security Considerations

Implementers **MUST** consider the security implications of exposing actions via MCP:

### Authentication

- **API Keys**: Simple string tokens sent via headers or metadata.
- **OAuth/OIDC**: For more complex scenarios with delegated permissions.
- **Mutual TLS**: For service-to-service communication with strong identity verification.
- **JWT Tokens**: For stateless authentication with claims.

Authentication information **SHOULD NOT** be included directly in the `params` object but rather in:
- HTTP headers (for HTTP transport)
- A dedicated `auth` field in metadata (for non-HTTP transports)

### Authorization

- **Action-level permissions**: Control which clients can invoke which actions.
- **Data-level permissions**: Filter results based on client identity.
- **Rate limiting per client**: Prevent abuse by limiting request frequency.
- **Audit logging**: Record all action invocations with client identity.

### Input Validation

- **Schema validation**: Validate all inputs against JSON Schema before processing.
- **Sanitization**: Clean input data to prevent injection attacks.
- **Size limits**: Enforce maximum sizes for all input parameters.
- **Whitelisting**: Restrict inputs to known-good values where possible.

### Rate Limiting

- **Implementation**: Use token bucket, fixed window, or sliding window algorithms.
- **Response format**: When rate limited, return a standard error:

```json
{
  "status": "error",
  "error": {
    "type": "RATE_LIMIT_EXCEEDED",
    "message": "Request rate limit exceeded",
    "details": {
      "retry_after_seconds": 30
    }
  }
}
```

- **Headers**: For HTTP transport, include standard headers:
  - `X-RateLimit-Limit`: Total allowed requests in period
  - `X-RateLimit-Remaining`: Remaining requests in current period
  - `X-RateLimit-Reset`: Time when the limit resets
  - `Retry-After`: Seconds to wait before retrying

---

## Versioning

### Semantic Versioning

MCP implementations **SHOULD** use semantic versioning (MAJOR.MINOR.PATCH):

- **MAJOR**: Breaking changes to the protocol
- **MINOR**: New features, backwards-compatible
- **PATCH**: Bug fixes, backwards-compatible

The protocol version **SHOULD** be communicated via:

- Metadata field: `"metadata": {"protocol_version": "1.0.0"}`
- HTTP header (for HTTP transport): `MCP-Version: 1.0.0`
- Version discovery endpoint or action

### Backward Compatibility

To maintain backward compatibility:

- **New fields**: May be added to requests/responses
- **Required fields**: Must not be removed or have their meaning changed
- **Action parameters**: May add new optional parameters, but not change or remove existing ones
- **Response structure**: May add new optional fields, but not change the meaning of existing ones

### Protocol Evolution

For evolving the protocol:

1. **Deprecation notices**: Before removing features, mark them as deprecated with notices
2. **Version negotiation**: Allow clients to specify acceptable versions
3. **Feature detection**: Use the schema discovery mechanism to detect supported features

Example version negotiation (HTTP):
```
Request: 
  MCP-Version: 2.0
  MCP-Min-Version: 1.0

Response:
  MCP-Version: 1.5
```

---

## Compliance Guidelines

To ensure interoperability and robustness, MCP implementations **SHOULD** adhere to the following:

- **Valid JSON:** Strictly adhere to JSON syntax for all requests and responses.
- **Required Fields:** Always include `action` and `params` in requests, and `status` in responses (`error` if status is "error").
- **Parameter Validation:** Servers **MUST** validate incoming `params` against the action's schema. Reject requests with invalid parameters using an appropriate error response.
- **Clear Errors:** Provide meaningful `error` objects when `status` is `"error"`, including a `type` and `message`.
- **Documented Actions:** Clearly document available actions, their parameters, expected output (`data` structure), and potential error types.
- **Transport Specifics:** Adhere to the conventions of the chosen transport layer (e.g., HTTP status codes, Content-Type headers).
- **Idempotency:** Where applicable, design actions to be idempotent (multiple identical requests have the same effect as one).
- **UTF-8 Encoding:** All text **MUST** be UTF-8 encoded.
- **Consistent Naming:** Follow naming conventions for actions and parameters.
- **Semantic Status Codes:** Use appropriate error types and HTTP status codes.
- **Documentation:** Provide complete API documentation, including examples.

---

## Implementation Guidance

### Language-Specific Recommendations

**Python:**
```python
import json
import jsonschema

# Define action schema
ADD_SCHEMA = {
    "type": "object",
    "properties": {
        "a": {"type": "number"},
        "b": {"type": "number"}
    },
    "required": ["a", "b"]
}

def handle_add(params):
    # Validate params
    jsonschema.validate(params, ADD_SCHEMA)
    
    # Process action
    result = params["a"] + params["b"]
    
    # Return success response
    return {
        "status": "success",
        "data": {"result": result}
    }

def handle_request(request_json):
    try:
        # Parse request
        request = json.loads(request_json)
        
        # Extract action and params
        action = request.get("action")
        params = request.get("params", {})
        
        if not action:
            return {
                "status": "error",
                "error": {
                    "type": "MISSING_ACTION",
                    "message": "No action specified in request"
                }
            }
        
        # Route to handler
        if action == "calculator.add":
            return handle_add(params)
        else:
            return {
                "status": "error",
                "error": {
                    "type": "ACTION_NOT_FOUND",
                    "message": f"Action '{action}' not supported"
                }
            }
            
    except json.JSONDecodeError:
        return {
            "status": "error",
            "error": {
                "type": "INVALID_JSON",
                "message": "Request contains invalid JSON"
            }
        }
    except jsonschema.exceptions.ValidationError as e:
        return {
            "status": "error",
            "error": {
                "type": "VALIDATION_ERROR",
                "message": str(e)
            }
        }
    except Exception as e:
        return {
            "status": "error",
            "error": {
                "type": "INTERNAL_ERROR",
                "message": "An unexpected error occurred"
            }
        }
```

**JavaScript/TypeScript:**
```typescript
import Ajv from 'ajv';

const ajv = new Ajv();

// Define schemas
const addSchema = {
  type: 'object',
  properties: {
    a: { type: 'number' },
    b: { type: 'number' }
  },
  required: ['a', 'b']
};

// Validate schemas
const validateAdd = ajv.compile(addSchema);

// Handle requests
async function handleRequest(requestData) {
  try {
    // Extract action and params
    const { action, params = {}, metadata = {} } = requestData;
    
    if (!action) {
      return {
        status: 'error',
        error: {
          type: 'MISSING_ACTION',
          message: 'No action specified in request'
        },
        metadata: { request_id: metadata.request_id }
      };
    }
    
    // Route to handler
    if (action === 'calculator.add') {
      return handleAdd(params, metadata);
    } else {
      return {
        status: 'error',
        error: {
          type: 'ACTION_NOT_FOUND',
          message: `Action '${action}' not supported`
        },
        metadata: { request_id: metadata.request_id }
      };
    }
  } catch (error) {
    return {
      status: 'error',
      error: {
        type: 'INTERNAL_ERROR',
        message: error.message || 'An unexpected error occurred'
      }
    };
  }
}

function handleAdd(params, metadata) {
  // Validate params
  if (!validateAdd(params)) {
    return {
      status: 'error',
      error: {
        type: 'VALIDATION_ERROR',
        message: ajv.errorsText(validateAdd.errors),
        details: validateAdd.errors
      },
      metadata: { request_id: metadata.request_id }
    };
  }
  
  // Process action
  const result = params.a + params.b;
  
  // Return success response
  return {
    status: 'success',
    data: { result },
    metadata: { request_id: metadata.request_id }
  };
}
```

### Testing Strategies

1. **Unit Testing:**
   - Test each action handler independently
   - Verify parameter validation logic
   - Test all error paths

2. **Integration Testing:**
   - Test full request/response flow
   - Verify correct handling of transport-specific details
   - Test authentication and authorization logic

3. **Specific Test Cases:**
   - Valid requests with minimal params
   - Valid requests with all optional params
   - Invalid JSON syntax
   - Missing required params
   - Invalid param types
   - Non-existent actions
   - Authentication failures
   - Rate limit testing
   - Large payload handling

4. **Performance Testing:**
   - Response time under normal load
   - Behavior under high concurrency
   - Memory usage with large payloads

### Performance Optimization

1. **Request Parsing:**
   - Use efficient JSON parsers
   - Consider schema compilation for validation
   - Cache compiled schemas

2. **Response Generation:**
   - Minimize allocations in hot paths
   - Use streaming for large responses
   - Consider response compression for HTTP

3. **Concurrency:**
   - Implement non-blocking I/O for high-concurrency scenarios
   - Use connection pooling for external dependencies
   - Consider worker pools for CPU-intensive operations

4. **Caching:**
   - Cache frequently requested data
   - Use ETags or conditional requests for HTTP
   - Implement expiration and invalidation policies

### Retry Strategies

Clients **SHOULD** implement retry logic for transient failures:

1. **Retryable Errors:**
   - Network connectivity issues
   - Rate limiting (429 Too Many Requests)
   - Server temporary unavailability (503 Service Unavailable)
   - Gateway timeouts (504 Gateway Timeout)

2. **Non-Retryable Errors:**
   - Authentication/authorization failures (401, 403)
   - Invalid requests (400 Bad Request)
   - Resource not found (404 Not Found)
   - Method not allowed (405 Method Not Allowed)

3. **Retry Algorithm:**
   - Start with a base delay (e.g., 1 second)
   - Use exponential backoff: delay = baseDelay * (2 ^ attemptNumber)
   - Add jitter (random variation) to prevent thundering herd problems
   - Set maximum retry count and maximum delay

4. **Timeout Handling:**
   - Set appropriate connect and read timeouts
   - Consider different timeout values for different actions
   - Implement client-side circuit breaking for consistently failing endpoints

---

## Summary

MCP provides a foundational protocol for structured communication in distributed AI systems. By defining clear message formats, supporting multiple transports, and encouraging good practices around error handling, schema definition, and security, it enables the development of modular, interoperable, and scalable applications involving LLMs, agents, and tools.

The protocol's simplicity and flexibility make it suitable for a wide range of use cases, from simple CLI tools to complex distributed systems. Its transport agnosticism allows integration with existing infrastructure, while its structured approach to actions and error handling promotes reliability and maintainability.

Implementers are encouraged to follow the guidelines in this specification while adapting to their specific requirements. By adhering to these conventions, developers can create an ecosystem of interoperable tools and agents that work together seamlessly, accelerating the development of AI-powered applications.
