package types

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ProtocolTypesTestSuite provides comprehensive tests for protocol-related types
type ProtocolTypesTestSuite struct {
	suite.Suite
}

func (s *ProtocolTypesTestSuite) TestTransportType() {
	s.Run("Constants", func() {
		s.Equal(TransportStdio, TransportType("stdio"))
		s.Equal(TransportHTTP, TransportType("http"))
	})

	s.Run("StringConversion", func() {
		s.Equal("stdio", string(TransportStdio))
		s.Equal("http", string(TransportHTTP))
	})

	s.Run("TypeAssignment", func() {
		var transport TransportType
		transport = TransportStdio
		s.Equal(TransportStdio, transport)

		transport = TransportHTTP
		s.Equal(TransportHTTP, transport)
	})

	s.Run("JSONSerialization", func() {
		type TestStruct struct {
			Transport TransportType `json:"transport"`
		}

		test := TestStruct{Transport: TransportStdio}
		data, err := json.Marshal(test)
		s.Require().NoError(err)
		s.Contains(string(data), "stdio")

		var decoded TestStruct
		err = json.Unmarshal(data, &decoded)
		s.Require().NoError(err)
		s.Equal(TransportStdio, decoded.Transport)
	})
}

func (s *ProtocolTypesTestSuite) TestMCPRequest() {
	s.Run("BasicStructure", func() {
		req := MCPRequest{
			JSONRPC: "2.0",
			ID:      "test-id",
			Method:  "test/method",
			Params:  map[string]interface{}{"key": "value"},
		}

		s.Equal("2.0", req.JSONRPC)
		s.Equal("test-id", req.ID)
		s.Equal("test/method", req.Method)
		s.NotNil(req.Params)
	})

	s.Run("DifferentIDTypes", func() {
		// String ID
		req1 := MCPRequest{ID: "string-id"}
		s.Equal("string-id", req1.ID)

		// Numeric ID
		req2 := MCPRequest{ID: 123}
		s.Equal(123, req2.ID)

		// Nil ID (notification)
		req3 := MCPRequest{ID: nil}
		s.Nil(req3.ID)
	})

	s.Run("JSONSerialization", func() {
		req := MCPRequest{
			JSONRPC: "2.0",
			ID:      42,
			Method:  "initialize",
			Params:  map[string]interface{}{"clientInfo": map[string]string{"name": "test"}},
		}

		data, err := json.Marshal(req)
		s.Require().NoError(err)
		s.NotEmpty(data)

		var decoded MCPRequest
		err = json.Unmarshal(data, &decoded)
		s.Require().NoError(err)

		s.Equal(req.JSONRPC, decoded.JSONRPC)
		s.InDelta(float64(42), decoded.ID, 0.0001) // JSON numbers decode as float64
		s.Equal(req.Method, decoded.Method)
		s.NotNil(decoded.Params)
	})

	s.Run("NoParams", func() {
		req := MCPRequest{
			JSONRPC: "2.0",
			ID:      "test",
			Method:  "ping",
		}

		data, err := json.Marshal(req)
		s.Require().NoError(err)

		// Should not include params field when nil/empty
		jsonStr := string(data)
		s.NotContains(jsonStr, "params")
	})
}

func (s *ProtocolTypesTestSuite) TestMCPResponse() {
	s.Run("SuccessResponse", func() {
		resp := MCPResponse{
			JSONRPC: "2.0",
			ID:      "test-id",
			Result:  map[string]interface{}{"success": true},
		}

		s.Equal("2.0", resp.JSONRPC)
		s.Equal("test-id", resp.ID)
		s.NotNil(resp.Result)
		s.Nil(resp.Error)
	})

	s.Run("ErrorResponse", func() {
		resp := MCPResponse{
			JSONRPC: "2.0",
			ID:      "test-id",
			Error: &MCPError{
				Code:    -32600,
				Message: "Invalid Request",
				Data:    "Additional error data",
			},
		}

		s.Equal("2.0", resp.JSONRPC)
		s.Equal("test-id", resp.ID)
		s.Nil(resp.Result)
		s.NotNil(resp.Error)
		s.Equal(-32600, resp.Error.Code)
		s.Equal("Invalid Request", resp.Error.Message)
	})

	s.Run("JSONSerialization", func() {
		resp := MCPResponse{
			JSONRPC: "2.0",
			ID:      123,
			Result:  map[string]string{"status": "ok"},
		}

		data, err := json.Marshal(resp)
		s.Require().NoError(err)
		s.NotEmpty(data)

		var decoded MCPResponse
		err = json.Unmarshal(data, &decoded)
		s.Require().NoError(err)

		s.Equal(resp.JSONRPC, decoded.JSONRPC)
		s.InDelta(float64(123), decoded.ID, 0.0001)
		s.NotNil(decoded.Result)
	})
}

func (s *ProtocolTypesTestSuite) TestMCPError() {
	s.Run("BasicError", func() {
		err := MCPError{
			Code:    -32700,
			Message: "Parse error",
		}

		s.Equal(-32700, err.Code)
		s.Equal("Parse error", err.Message)
		s.Nil(err.Data)
	})

	s.Run("ErrorWithData", func() {
		err := MCPError{
			Code:    -32602,
			Message: "Invalid params",
			Data:    map[string]interface{}{"expected": "string", "got": "number"},
		}

		s.Equal(-32602, err.Code)
		s.Equal("Invalid params", err.Message)
		s.NotNil(err.Data)
	})

	s.Run("JSONSerialization", func() {
		err := MCPError{
			Code:    -32601,
			Message: "Method not found",
			Data:    "method 'unknown' not supported",
		}

		data, errMarshal := json.Marshal(err)
		s.Require().NoError(errMarshal)
		s.NotEmpty(data)

		var decoded MCPError
		errUnmarshal := json.Unmarshal(data, &decoded)
		s.Require().NoError(errUnmarshal)

		s.Equal(err.Code, decoded.Code)
		s.Equal(err.Message, decoded.Message)
		s.Equal(err.Data, decoded.Data)
	})
}

func (s *ProtocolTypesTestSuite) TestToolCallParams() {
	s.Run("BasicStructure", func() {
		params := ToolCallParams{
			Name: "test_tool",
			Arguments: map[string]interface{}{
				"param1": "value1",
				"param2": 42,
			},
		}

		s.Equal("test_tool", params.Name)
		s.NotNil(params.Arguments)
		s.Equal("value1", params.Arguments["param1"])
		s.Equal(42, params.Arguments["param2"])
	})

	s.Run("NoArguments", func() {
		params := ToolCallParams{
			Name: "simple_tool",
		}

		s.Equal("simple_tool", params.Name)
		s.Nil(params.Arguments)
	})

	s.Run("JSONSerialization", func() {
		params := ToolCallParams{
			Name: "complex_tool",
			Arguments: map[string]interface{}{
				"string_param":  "test",
				"number_param":  123.45,
				"boolean_param": true,
				"array_param":   []interface{}{"a", "b", "c"},
			},
		}

		data, err := json.Marshal(params)
		s.Require().NoError(err)
		s.NotEmpty(data)

		var decoded ToolCallParams
		err = json.Unmarshal(data, &decoded)
		s.Require().NoError(err)

		s.Equal(params.Name, decoded.Name)
		s.Equal(params.Arguments["string_param"], decoded.Arguments["string_param"])
		s.Equal(params.Arguments["boolean_param"], decoded.Arguments["boolean_param"])
	})
}

func (s *ProtocolTypesTestSuite) TestToolCallResult() {
	s.Run("SuccessResult", func() {
		result := ToolCallResult{
			Content: []Content{
				{Type: "text", Text: "Success"},
			},
			IsError: false,
		}

		s.Len(result.Content, 1)
		s.Equal("text", result.Content[0].Type)
		s.Equal("Success", result.Content[0].Text)
		s.False(result.IsError)
	})

	s.Run("ErrorResult", func() {
		result := ToolCallResult{
			Content: []Content{
				{Type: "text", Text: "Error occurred"},
			},
			IsError: true,
		}

		s.Len(result.Content, 1)
		s.Equal("Error occurred", result.Content[0].Text)
		s.True(result.IsError)
	})

	s.Run("MultipleContent", func() {
		result := ToolCallResult{
			Content: []Content{
				{Type: "text", Text: "First part"},
				{Type: "resource", Resource: "file://test.txt", MimeType: "text/plain"},
			},
		}

		s.Len(result.Content, 2)
		s.Equal("text", result.Content[0].Type)
		s.Equal("resource", result.Content[1].Type)
		s.Equal("file://test.txt", result.Content[1].Resource)
		s.Equal("text/plain", result.Content[1].MimeType)
	})
}

func (s *ProtocolTypesTestSuite) TestContent() {
	s.Run("TextContent", func() {
		content := Content{
			Type: "text",
			Text: "Hello, world!",
		}

		s.Equal("text", content.Type)
		s.Equal("Hello, world!", content.Text)
		s.Empty(content.Resource)
		s.Empty(content.MimeType)
	})

	s.Run("ResourceContent", func() {
		content := Content{
			Type:     "resource",
			Resource: "file://example.json",
			MimeType: "application/json",
		}

		s.Equal("resource", content.Type)
		s.Empty(content.Text)
		s.Equal("file://example.json", content.Resource)
		s.Equal("application/json", content.MimeType)
	})

	s.Run("JSONSerialization", func() {
		content := Content{
			Type:     "text",
			Text:     "Sample text",
			MimeType: "text/plain",
		}

		data, err := json.Marshal(content)
		s.Require().NoError(err)
		s.NotEmpty(data)

		var decoded Content
		err = json.Unmarshal(data, &decoded)
		s.Require().NoError(err)

		s.Equal(content.Type, decoded.Type)
		s.Equal(content.Text, decoded.Text)
		s.Equal(content.MimeType, decoded.MimeType)
	})
}

func (s *ProtocolTypesTestSuite) TestInitializeResult() {
	s.Run("CompleteInitialization", func() {
		result := InitializeResult{
			ProtocolVersion: "2024-11-05",
			ServerInfo: ServerInfo{
				Name:    "go-invoice-mcp",
				Version: "1.0.0",
			},
			Capabilities: Capabilities{
				Tools: &ToolsCapability{
					ListChanged: true,
				},
				Resources: &ResourcesCapability{
					Subscribe:   true,
					ListChanged: true,
				},
			},
		}

		s.Equal("2024-11-05", result.ProtocolVersion)
		s.Equal("go-invoice-mcp", result.ServerInfo.Name)
		s.Equal("1.0.0", result.ServerInfo.Version)
		s.NotNil(result.Capabilities.Tools)
		s.True(result.Capabilities.Tools.ListChanged)
		s.NotNil(result.Capabilities.Resources)
		s.True(result.Capabilities.Resources.Subscribe)
	})

	s.Run("MinimalInitialization", func() {
		result := InitializeResult{
			ProtocolVersion: "2024-11-05",
			ServerInfo: ServerInfo{
				Name:    "minimal-server",
				Version: "0.1.0",
			},
			Capabilities: Capabilities{},
		}

		s.Equal("2024-11-05", result.ProtocolVersion)
		s.Equal("minimal-server", result.ServerInfo.Name)
		s.Nil(result.Capabilities.Tools)
		s.Nil(result.Capabilities.Resources)
	})
}

func (s *ProtocolTypesTestSuite) TestCapabilities() {
	s.Run("AllCapabilities", func() {
		caps := Capabilities{
			Tools: &ToolsCapability{
				ListChanged: true,
			},
			Resources: &ResourcesCapability{
				Subscribe:   true,
				ListChanged: false,
			},
			Prompts: &PromptsCapability{
				ListChanged: true,
			},
			Logging: &LoggingCapability{
				Level: "debug",
			},
		}

		s.NotNil(caps.Tools)
		s.True(caps.Tools.ListChanged)
		s.NotNil(caps.Resources)
		s.True(caps.Resources.Subscribe)
		s.False(caps.Resources.ListChanged)
		s.NotNil(caps.Prompts)
		s.True(caps.Prompts.ListChanged)
		s.NotNil(caps.Logging)
		s.Equal("debug", caps.Logging.Level)
	})

	s.Run("NoCapabilities", func() {
		caps := Capabilities{}
		s.Nil(caps.Tools)
		s.Nil(caps.Resources)
		s.Nil(caps.Prompts)
		s.Nil(caps.Logging)
	})
}

func (s *ProtocolTypesTestSuite) TestTool() {
	s.Run("BasicTool", func() {
		tool := Tool{
			Name:        "test_tool",
			Description: "A test tool",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"param": map[string]interface{}{
						"type": "string",
					},
				},
			},
		}

		s.Equal("test_tool", tool.Name)
		s.Equal("A test tool", tool.Description)
		s.NotNil(tool.InputSchema)
		s.Equal("object", tool.InputSchema["type"])
	})

	s.Run("JSONSerialization", func() {
		tool := Tool{
			Name:        "json_tool",
			Description: "Tool for JSON testing",
			InputSchema: map[string]interface{}{
				"type":     "object",
				"required": []string{"name"},
			},
		}

		data, err := json.Marshal(tool)
		s.Require().NoError(err)
		s.NotEmpty(data)

		var decoded Tool
		err = json.Unmarshal(data, &decoded)
		s.Require().NoError(err)

		s.Equal(tool.Name, decoded.Name)
		s.Equal(tool.Description, decoded.Description)
		s.NotNil(decoded.InputSchema)
	})
}

func (s *ProtocolTypesTestSuite) TestInitializeParams() {
	s.Run("BasicParams", func() {
		params := InitializeParams{
			ProtocolVersion: "2024-11-05",
			Capabilities: map[string]interface{}{
				"roots": map[string]interface{}{
					"listChanged": true,
				},
			},
			ClientInfo: ClientInfo{
				Name:    "test-client",
				Version: "1.0.0",
			},
		}

		s.Equal("2024-11-05", params.ProtocolVersion)
		s.NotNil(params.Capabilities)
		s.Equal("test-client", params.ClientInfo.Name)
		s.Equal("1.0.0", params.ClientInfo.Version)
	})
}

func (s *ProtocolTypesTestSuite) TestInterfaces() {
	s.Run("InterfaceDeclaration", func() {
		// Test that interfaces are properly declared with expected methods
		// We can't instantiate interfaces directly, but we can check they exist
		s.Run("Server", func() {
			var server Server
			s.Nil(server)
			// Methods: Start, Shutdown, HandleRequest
		})

		s.Run("CLIBridge", func() {
			var bridge CLIBridge
			s.Nil(bridge)
			// Methods: ExecuteCommand, ValidateCommand, GetAllowedCommands
		})

		s.Run("Logger", func() {
			var logger Logger
			s.Nil(logger)
			// Methods: Debug, Info, Warn, Error
		})

		s.Run("CommandValidator", func() {
			var validator CommandValidator
			s.Nil(validator)
			// Methods: ValidateCommand, IsCommandAllowed
		})

		s.Run("FileHandler", func() {
			var handler FileHandler
			s.Nil(handler)
			// Methods: PrepareWorkspace, ValidatePath
		})

		s.Run("MCPHandler", func() {
			var handler MCPHandler
			s.Nil(handler)
			// Methods: HandleRequest
		})
	})
}

// TestProtocolTypesTestSuite runs the complete protocol types test suite
func TestProtocolTypesTestSuite(t *testing.T) {
	suite.Run(t, new(ProtocolTypesTestSuite))
}

// Benchmark tests for protocol types
func BenchmarkMCPRequestMarshal(b *testing.B) {
	req := MCPRequest{
		JSONRPC: "2.0",
		ID:      123,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name": "test_tool",
			"arguments": map[string]interface{}{
				"param1": "value1",
				"param2": 42,
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMCPResponseMarshal(b *testing.B) {
	resp := MCPResponse{
		JSONRPC: "2.0",
		ID:      123,
		Result: map[string]interface{}{
			"content": []Content{
				{Type: "text", Text: "Result"},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(resp)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Unit tests for specific behaviors
func TestProtocolTypes_Specific(t *testing.T) {
	t.Run("TransportTypeValidation", func(t *testing.T) {
		validTransports := []TransportType{TransportStdio, TransportHTTP}
		for _, transport := range validTransports {
			assert.NotEmpty(t, string(transport))
		}

		// Custom transport type
		custom := TransportType("custom")
		assert.Equal(t, "custom", string(custom))
	})

	t.Run("MCPErrorCodes", func(t *testing.T) {
		// Standard JSON-RPC error codes
		standardCodes := []int{-32700, -32600, -32601, -32602, -32603}
		for _, code := range standardCodes {
			err := MCPError{Code: code, Message: "test"}
			assert.Equal(t, code, err.Code)
		}
	})

	t.Run("ContentTypes", func(t *testing.T) {
		types := []string{"text", "resource", "image", "audio"}
		for _, contentType := range types {
			content := Content{Type: contentType}
			assert.Equal(t, contentType, content.Type)
		}
	})
}

// Edge case tests
func TestProtocolTypes_EdgeCases(t *testing.T) {
	t.Run("EmptyValues", func(t *testing.T) {
		// Empty request
		req := MCPRequest{}
		data, err := json.Marshal(req)
		require.NoError(t, err)
		assert.NotEmpty(t, data)

		// Empty response
		resp := MCPResponse{}
		data, err = json.Marshal(resp)
		require.NoError(t, err)
		assert.NotEmpty(t, data)
	})

	t.Run("NilPointers", func(t *testing.T) {
		// Response with nil error
		resp := MCPResponse{Error: nil}
		assert.Nil(t, resp.Error)

		// Capabilities with nil pointers
		caps := Capabilities{
			Tools:     nil,
			Resources: nil,
			Prompts:   nil,
			Logging:   nil,
		}
		assert.Nil(t, caps.Tools)
		assert.Nil(t, caps.Resources)
		assert.Nil(t, caps.Prompts)
		assert.Nil(t, caps.Logging)
	})

	t.Run("JSONOmitEmpty", func(t *testing.T) {
		// Content with empty fields
		content := Content{Type: "text", Text: "hello"}
		data, err := json.Marshal(content)
		require.NoError(t, err)

		jsonStr := string(data)
		assert.Contains(t, jsonStr, "text")
		assert.Contains(t, jsonStr, "hello")
		// Should not contain empty fields
		assert.NotContains(t, jsonStr, "resource")
		assert.NotContains(t, jsonStr, "mimeType")
	})

	t.Run("ComplexJSONHandling", func(t *testing.T) {
		// Test with nested complex structures
		params := map[string]interface{}{
			"nested": map[string]interface{}{
				"array":  []interface{}{1, "two", true},
				"object": map[string]interface{}{"key": "value"},
			},
		}

		req := MCPRequest{
			JSONRPC: "2.0",
			ID:      "complex",
			Method:  "test",
			Params:  params,
		}

		data, err := json.Marshal(req)
		require.NoError(t, err)

		var decoded MCPRequest
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, req.JSONRPC, decoded.JSONRPC)
		assert.Equal(t, req.ID, decoded.ID)
		assert.Equal(t, req.Method, decoded.Method)
		assert.NotNil(t, decoded.Params)
	})
}

// Context integration tests
func TestProtocolTypes_Context(t *testing.T) {
	t.Run("ContextUsage", func(t *testing.T) {
		// Test that context can be used with types (even though types don't directly use context)
		ctx := context.Background()
		type contextKey string
		const requestIDKey contextKey = "request_id"
		ctx = context.WithValue(ctx, requestIDKey, "test-123")

		// Simulate usage in interface methods
		assert.NotNil(t, ctx)
		assert.Equal(t, "test-123", ctx.Value(requestIDKey))
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		select {
		case <-ctx.Done():
			require.Error(t, ctx.Err())
		default:
			t.Fatal("Context should be canceled")
		}
	})
}
