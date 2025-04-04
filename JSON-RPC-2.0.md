# JSON-RPC 2.0 Specification

## Overview
JSON-RPC is a stateless, light-weight remote procedure call (RPC) protocol that uses JSON ([RFC 4627](http://www.ietf.org/rfc/rfc4627.txt)) as data format. This specification defines JSON-RPC version 2.0.

## Conventions
The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT", "SHOULD", "SHOULD NOT", "RECOMMENDED", "MAY", and "OPTIONAL" in this document are to be interpreted as described in [RFC 2119](http://www.ietf.org/rfc/rfc2119.txt).

Since JSON-RPC utilizes JSON, it has the same encoding considerations as described in the JSON specification.

## 1. Request Object

A Request object has the following members:

| Member    | Type          | Description |
|-----------|---------------|-------------|
| jsonrpc   | string        | A string specifying the version of the JSON-RPC protocol. MUST be exactly "2.0". |
| method    | string        | A string containing the name of the method to be invoked. Method names that begin with "rpc." are reserved for system extensions and MUST NOT be used for anything else. |
| params    | array/object  | A structured value that holds the parameter values to be used during the invocation of the method. This member MAY be omitted. |
| id        | string/number/null | An identifier established by the client that MUST contain a string, number, or NULL value if included. If it is not included it is assumed to be a notification. The value SHOULD normally not be Null [1] and Numbers SHOULD NOT contain fractional parts [2]. |

### Notification
A Notification is a Request object without an "id" member. Notifications are intended for one-way communication, and the Server MUST NOT reply to a Notification, including those within a batch request. Clients should not expect any acknowledgment or error reporting from the server.

### Parameter Structures
Parameters can be specified by-position, by-name, or omitted entirely.

- By-position: params MUST be an array with elements in the expected order.
- By-name: params MUST be an object with member names that match the expected parameter names.
- Omitted: params MAY be omitted entirely if the method requires no parameters.

## 2. Response Object

When a rpc call is made, the Server MUST reply with a Response, except for Notifications. The Response is expressed as a single JSON Object with the following members:

| Member    | Type          | Description |
|-----------|---------------|-------------|
| jsonrpc   | string        | A string specifying the version of the JSON-RPC protocol. MUST be exactly "2.0". |
| result    | any           | This member is REQUIRED on success. This member MUST NOT exist if there was an error invoking the method. The value is determined by the method invoked on the Server. |
| error     | object        | This member is REQUIRED on error. This member MUST NOT exist if there was no error triggered during invocation. |
| id        | string/number/null | This member is REQUIRED. It MUST be the same as the value of the id member in the Request Object. If there was an error in detecting the id in the Request object (e.g. Parse error/Invalid Request), it MUST be Null. |

Either the result or error member MUST be included, but both members MUST NOT be included in the same response object.

### Error Object

When a rpc call encounters an error, the Response Object MUST contain the error member with a value that is a JSON Object with the following members:

| Member    | Type          | Description |
|-----------|---------------|-------------|
| code      | number        | A Number that indicates the error type that occurred. |
| message   | string        | A String providing a short description of the error. |
| data      | any           | A Primitive or Structured value that contains additional information about the error. This may be omitted. The value of this member is defined by the Server (e.g. detailed error information, nested errors etc.). |

#### Error Codes

The error codes from and including -32768 to -32000 are reserved for pre-defined errors:

| Code      | Message       | Meaning |
|-----------|---------------|---------|
| -32700    | Parse error   | Invalid JSON was received by the server. |
| -32600    | Invalid Request | The JSON sent is not a valid Request object. |
| -32601    | Method not found | The method does not exist / is not available. |
| -32602    | Invalid params | Invalid method parameter(s). |
| -32603    | Internal error | Internal JSON-RPC error. |
| -32000 to -32099 | Server error | Reserved for implementation-defined server-errors. |

The remainder of the space is available for application defined errors.

## 3. Batch Processing

Multiple Request objects can be sent to the server as an array. The server should respond with an array of Response objects.

A Request object that is a Notification will not be included in the Response array.

If the batch is empty, the server MUST return an Invalid Request error with the following structure:
```json
{"jsonrpc": "2.0", "error": {"code": -32600, "message": "Invalid Request"}, "id": null}
```

If the batch rpc call itself fails to be recognized as valid JSON or as an Array with at least one value, the response from the Server MUST be a single Response object indicating a Parse error or Invalid Request error accordingly.

If there are no Response objects contained within the batch (i.e., all notifications), the server SHOULD NOT respond at all.

## 4. Examples

### Request Examples

#### A notification (no response expected):
```json
{"jsonrpc": "2.0", "method": "update", "params": [1, 2, 3, 4, 5]}
```

#### A request with positional parameters:
```json
{"jsonrpc": "2.0", "method": "subtract", "params": [42, 23], "id": 1}
```

#### A request with named parameters:
```json
{"jsonrpc": "2.0", "method": "subtract", "params": {"subtrahend": 23, "minuend": 42}, "id": 3}
```

### Response Examples

#### Response to a successful request:
```json
{"jsonrpc": "2.0", "result": 19, "id": 1}
```

#### Response to a request that resulted in an error:
```json
{"jsonrpc": "2.0", "error": {"code": -32601, "message": "Method not found"}, "id": "1"}
```

#### Response to a notification (there should be none)
```
[no response]
```

### Batch Examples

#### Batch request:
```json
[
  {"jsonrpc": "2.0", "method": "sum", "params": [1,2,4], "id": "1"},
  {"jsonrpc": "2.0", "method": "notify_hello", "params": [7]},
  {"jsonrpc": "2.0", "method": "subtract", "params": [42,23], "id": "2"},
  {"jsonrpc": "2.0", "method": "foo.get", "params": {"name": "myself"}, "id": "5"}
]
```

#### Batch response:
```json
[
  {"jsonrpc": "2.0", "result": 7, "id": "1"},
  {"jsonrpc": "2.0", "result": 19, "id": "2"},
  {"jsonrpc": "2.0", "error": {"code": -32601, "message": "Method not found"}, "id": "5"}
]
```

### Additional Examples

#### Empty batch request:
```json
[]
```

#### Response to empty batch request:
```json
{"jsonrpc": "2.0", "error": {"code": -32600, "message": "Invalid Request"}, "id": null}
```

#### Invalid JSON request:
```json
{"jsonrpc": "2.0", "method": "subtract", "params": [42, 23], "id": 1
```

#### Response to invalid JSON request:
```json
{"jsonrpc": "2.0", "error": {"code": -32700, "message": "Parse error"}, "id": null}
```

#### Batch request with only notifications:
```json
[
  {"jsonrpc": "2.0", "method": "notify_hello", "params": [7]},
  {"jsonrpc": "2.0", "method": "notify_update", "params": [1,2,3]}
]
```

#### Response to batch request with only notifications:
```
[no response]
```

## 5. Extensions

The JSON-RPC 2.0 specification deliberately omits some features to keep it simple and flexible. However, implementers MAY choose to extend the protocol with additional features while maintaining compatibility with the core specification.

Common extensions include:
- Transport-specific binding details
- Authentication mechanisms
- Bi-directional communication patterns
- Service discovery
- Introspection capabilities

## 6. Implementation Considerations

### Character Encoding
JSON-RPC 2.0 uses JSON, which MUST be encoded in UTF-8 when transmitted over a network. Deviations from UTF-8 encoding may lead to interoperability issues.

### Security
The protocol itself does not provide any security features. Implementers SHOULD use appropriate transport layer security mechanisms, such as TLS or HTTPS, when sensitive data is being exchanged.

### Performance
For high-performance applications, implementers should consider:
- Minimizing the size of messages
- Using efficient JSON parsers
- Implementing connection pooling when appropriate
- Considering binary alternatives for very large datasets

## References

1. [JSON-RPC 2.0 Official Specification](https://www.jsonrpc.org/specification)
2. [JSON Specification (RFC 4627)](http://www.ietf.org/rfc/rfc4627.txt) - Note [1]: It's desirable that Null isn't used as an id for Responses which are Notifications. Note [2]: Fractional parts may not parse correctly.
3. [Keyword Guidelines (RFC 2119)](http://www.ietf.org/rfc/rfc2119.txt)
