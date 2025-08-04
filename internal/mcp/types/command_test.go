package types

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// CommandTypesTestSuite provides comprehensive tests for command-related types
type CommandTypesTestSuite struct {
	suite.Suite
}

func (s *CommandTypesTestSuite) TestCommandRequest() {
	s.Run("BasicStructure", func() {
		req := CommandRequest{
			Command:    "go",
			Args:       []string{"version"},
			WorkingDir: "/tmp",
			Env:        map[string]string{"GO_ENV": "test"},
			Timeout:    30 * time.Second,
			ExpectJSON: true,
			InputFiles: []string{"input.txt"},
		}

		s.Equal("go", req.Command)
		s.Equal([]string{"version"}, req.Args)
		s.Equal("/tmp", req.WorkingDir)
		s.Equal(map[string]string{"GO_ENV": "test"}, req.Env)
		s.Equal(30*time.Second, req.Timeout)
		s.True(req.ExpectJSON)
		s.Equal([]string{"input.txt"}, req.InputFiles)
	})

	s.Run("MinimalRequest", func() {
		req := CommandRequest{
			Command: "echo",
			Args:    []string{"hello"},
		}

		s.Equal("echo", req.Command)
		s.Equal([]string{"hello"}, req.Args)
		s.Empty(req.WorkingDir)
		s.Empty(req.Env)
		s.Equal(time.Duration(0), req.Timeout)
		s.False(req.ExpectJSON)
		s.Empty(req.InputFiles)
	})

	s.Run("JSONSerialization", func() {
		req := CommandRequest{
			Command:    "test",
			Args:       []string{"arg1", "arg2"},
			WorkingDir: "/path",
			Env:        map[string]string{"KEY": "value"},
			Timeout:    time.Minute,
			ExpectJSON: true,
			InputFiles: []string{"file1.txt", "file2.txt"},
		}

		data, err := json.Marshal(req)
		s.Require().NoError(err)
		s.NotEmpty(data)

		var decoded CommandRequest
		err = json.Unmarshal(data, &decoded)
		s.Require().NoError(err)

		s.Equal(req.Command, decoded.Command)
		s.Equal(req.Args, decoded.Args)
		s.Equal(req.WorkingDir, decoded.WorkingDir)
		s.Equal(req.Env, decoded.Env)
		s.Equal(req.Timeout, decoded.Timeout)
		s.Equal(req.ExpectJSON, decoded.ExpectJSON)
		s.Equal(req.InputFiles, decoded.InputFiles)
	})

	s.Run("EmptyCommand", func() {
		req := CommandRequest{}
		s.Empty(req.Command)
		s.Empty(req.Args)
	})

	s.Run("NilArgs", func() {
		req := CommandRequest{
			Command: "test",
		}
		s.Empty(req.Args)
	})

	s.Run("NilEnv", func() {
		req := CommandRequest{
			Command: "test",
		}
		s.Empty(req.Env)
	})

	s.Run("NilInputFiles", func() {
		req := CommandRequest{
			Command: "test",
		}
		s.Empty(req.InputFiles)
	})
}

func (s *CommandTypesTestSuite) TestCommandResponse() {
	s.Run("BasicStructure", func() {
		resp := CommandResponse{
			ExitCode: 0,
			Stdout:   "output",
			Stderr:   "error",
			Duration: time.Second,
			Error:    "test error",
			Files:    []string{"output.txt"},
		}

		s.Equal(0, resp.ExitCode)
		s.Equal("output", resp.Stdout)
		s.Equal("error", resp.Stderr)
		s.Equal(time.Second, resp.Duration)
		s.Equal("test error", resp.Error)
		s.Equal([]string{"output.txt"}, resp.Files)
	})

	s.Run("SuccessResponse", func() {
		resp := CommandResponse{
			ExitCode: 0,
			Stdout:   "success",
			Duration: 500 * time.Millisecond,
		}

		s.Equal(0, resp.ExitCode)
		s.Equal("success", resp.Stdout)
		s.Empty(resp.Stderr)
		s.Equal(500*time.Millisecond, resp.Duration)
		s.Empty(resp.Error)
		s.Empty(resp.Files)
	})

	s.Run("ErrorResponse", func() {
		resp := CommandResponse{
			ExitCode: 1,
			Stderr:   "command failed",
			Duration: time.Millisecond,
			Error:    "execution error",
		}

		s.Equal(1, resp.ExitCode)
		s.Empty(resp.Stdout)
		s.Equal("command failed", resp.Stderr)
		s.Equal(time.Millisecond, resp.Duration)
		s.Equal("execution error", resp.Error)
		s.Empty(resp.Files)
	})

	s.Run("JSONSerialization", func() {
		resp := CommandResponse{
			ExitCode: 2,
			Stdout:   "standard output",
			Stderr:   "standard error",
			Duration: 2 * time.Second,
			Error:    "some error",
			Files:    []string{"result1.txt", "result2.json"},
		}

		data, err := json.Marshal(resp)
		s.Require().NoError(err)
		s.NotEmpty(data)

		var decoded CommandResponse
		err = json.Unmarshal(data, &decoded)
		s.Require().NoError(err)

		s.Equal(resp.ExitCode, decoded.ExitCode)
		s.Equal(resp.Stdout, decoded.Stdout)
		s.Equal(resp.Stderr, decoded.Stderr)
		s.Equal(resp.Duration, decoded.Duration)
		s.Equal(resp.Error, decoded.Error)
		s.Equal(resp.Files, decoded.Files)
	})

	s.Run("ZeroValues", func() {
		resp := CommandResponse{}
		s.Equal(0, resp.ExitCode)
		s.Empty(resp.Stdout)
		s.Empty(resp.Stderr)
		s.Equal(time.Duration(0), resp.Duration)
		s.Empty(resp.Error)
		s.Empty(resp.Files)
	})
}

func (s *CommandTypesTestSuite) TestCommandRequestValidation() {
	s.Run("ValidCommands", func() {
		validCommands := []CommandRequest{
			{Command: "ls", Args: []string{"-la"}},
			{Command: "go", Args: []string{"test", "./..."}},
			{Command: "echo", Args: []string{"hello world"}},
			{Command: "pwd"},
		}

		for _, cmd := range validCommands {
			s.NotEmpty(cmd.Command, "Command should not be empty")
		}
	})

	s.Run("EdgeCases", func() {
		// Empty command
		req := CommandRequest{}
		s.Empty(req.Command)

		// Command with special characters
		req2 := CommandRequest{Command: "echo", Args: []string{"hello 'world' \"test\""}}
		s.NotEmpty(req2.Command)
		s.Contains(req2.Args[0], "'")
		s.Contains(req2.Args[0], "\"")

		// Long timeout
		req3 := CommandRequest{Timeout: 24 * time.Hour}
		s.Equal(24*time.Hour, req3.Timeout)

		// Many args
		manyArgs := make([]string, 100)
		for i := range manyArgs {
			manyArgs[i] = "arg"
		}
		req4 := CommandRequest{Command: "test", Args: manyArgs}
		s.Len(req4.Args, 100)
	})
}

func (s *CommandTypesTestSuite) TestCommandResponseAnalysis() {
	s.Run("IsSuccess", func() {
		successResp := CommandResponse{ExitCode: 0}
		s.Equal(0, successResp.ExitCode)

		errorResp := CommandResponse{ExitCode: 1}
		s.NotEqual(0, errorResp.ExitCode)
	})

	s.Run("HasOutput", func() {
		withOutput := CommandResponse{Stdout: "some output"}
		s.NotEmpty(withOutput.Stdout)

		withError := CommandResponse{Stderr: "some error"}
		s.NotEmpty(withError.Stderr)

		empty := CommandResponse{}
		s.Empty(empty.Stdout)
		s.Empty(empty.Stderr)
	})

	s.Run("HasFiles", func() {
		withFiles := CommandResponse{Files: []string{"file1", "file2"}}
		s.NotEmpty(withFiles.Files)

		noFiles := CommandResponse{}
		s.Empty(noFiles.Files)
	})
}

func (s *CommandTypesTestSuite) TestJSONOmitEmpty() {
	s.Run("OmitEmptyFields", func() {
		req := CommandRequest{
			Command: "test",
			Args:    []string{"arg"},
		}

		data, err := json.Marshal(req)
		s.Require().NoError(err)

		// Should not contain omitempty fields when they're empty
		jsonStr := string(data)
		s.NotContains(jsonStr, "workingDir")
		s.NotContains(jsonStr, "env")
		s.NotContains(jsonStr, "timeout")
		s.NotContains(jsonStr, "expectJSON")
		s.NotContains(jsonStr, "inputFiles")
	})

	s.Run("IncludeNonEmptyFields", func() {
		req := CommandRequest{
			Command:    "test",
			WorkingDir: "/tmp",
			ExpectJSON: true,
		}

		data, err := json.Marshal(req)
		s.Require().NoError(err)

		jsonStr := string(data)
		s.Contains(jsonStr, "workingDir")
		s.Contains(jsonStr, "expectJSON")
	})
}

// TestCommandTypesTestSuite runs the complete command types test suite
func TestCommandTypesTestSuite(t *testing.T) {
	suite.Run(t, new(CommandTypesTestSuite))
}

// Benchmark tests for command types
func BenchmarkCommandRequestMarshal(b *testing.B) {
	req := CommandRequest{
		Command:    "go",
		Args:       []string{"test", "./..."},
		WorkingDir: "/tmp",
		Env:        map[string]string{"GO_ENV": "test"},
		Timeout:    30 * time.Second,
		ExpectJSON: true,
		InputFiles: []string{"input.txt"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCommandResponseMarshal(b *testing.B) {
	resp := CommandResponse{
		ExitCode: 0,
		Stdout:   "output",
		Stderr:   "error",
		Duration: time.Second,
		Files:    []string{"output.txt"},
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
func TestCommandRequest_Specific(t *testing.T) {
	t.Run("TimeoutHandling", func(t *testing.T) {
		req := CommandRequest{Timeout: 5 * time.Second}
		assert.Equal(t, 5*time.Second, req.Timeout)

		// Zero timeout
		req2 := CommandRequest{}
		assert.Equal(t, time.Duration(0), req2.Timeout)
	})

	t.Run("ArgsHandling", func(t *testing.T) {
		// Nil args
		req := CommandRequest{Command: "test"}
		assert.Nil(t, req.Args)

		// Empty args
		req2 := CommandRequest{Command: "test", Args: []string{}}
		assert.NotNil(t, req2.Args)
		assert.Empty(t, req2.Args)

		// Args with empty strings
		req3 := CommandRequest{Command: "test", Args: []string{"", "arg", ""}}
		assert.Len(t, req3.Args, 3)
		assert.Empty(t, req3.Args[0])
		assert.Equal(t, "arg", req3.Args[1])
		assert.Empty(t, req3.Args[2])
	})
}

func TestCommandResponse_Specific(t *testing.T) {
	t.Run("ExitCodeHandling", func(t *testing.T) {
		// Success
		resp := CommandResponse{ExitCode: 0}
		assert.Equal(t, 0, resp.ExitCode)

		// Various error codes
		codes := []int{1, 2, 127, 128, 255}
		for _, code := range codes {
			resp := CommandResponse{ExitCode: code}
			assert.Equal(t, code, resp.ExitCode)
		}
	})

	t.Run("DurationHandling", func(t *testing.T) {
		durations := []time.Duration{
			0,
			time.Nanosecond,
			time.Microsecond,
			time.Millisecond,
			time.Second,
			time.Minute,
			time.Hour,
		}

		for _, duration := range durations {
			resp := CommandResponse{Duration: duration}
			assert.Equal(t, duration, resp.Duration)
		}
	})

	t.Run("OutputHandling", func(t *testing.T) {
		// Large output
		largeOutput := string(make([]byte, 10000))
		resp := CommandResponse{Stdout: largeOutput}
		assert.Len(t, resp.Stdout, 10000)

		// Unicode output
		unicodeOutput := "Hello World with Unicode"
		resp2 := CommandResponse{Stdout: unicodeOutput}
		assert.Equal(t, unicodeOutput, resp2.Stdout)

		// Binary-like output (should still work as string)
		binaryLike := "\x00\x01\x02\xFF"
		resp3 := CommandResponse{Stdout: binaryLike}
		assert.Equal(t, binaryLike, resp3.Stdout)
	})
}

// Edge case tests
func TestCommandTypes_EdgeCases(t *testing.T) {
	t.Run("JSONEdgeCases", func(t *testing.T) {
		// Request with nil map
		req := CommandRequest{Command: "test", Env: nil}
		data, err := json.Marshal(req)
		require.NoError(t, err)

		var decoded CommandRequest
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)
		assert.Nil(t, decoded.Env)

		// Response with nil slice
		resp := CommandResponse{Files: nil}
		data, err = json.Marshal(resp)
		require.NoError(t, err)

		var decodedResp CommandResponse
		err = json.Unmarshal(data, &decodedResp)
		require.NoError(t, err)
		assert.Nil(t, decodedResp.Files)
	})

	t.Run("StructEquality", func(t *testing.T) {
		req1 := CommandRequest{Command: "test", Args: []string{"arg"}}
		req2 := CommandRequest{Command: "test", Args: []string{"arg"}}

		// Go structs with same field values are equal with ==
		// but slices and maps need deep comparison
		assert.Equal(t, req1.Command, req2.Command)
		assert.Equal(t, req1.Args, req2.Args)

		resp1 := CommandResponse{ExitCode: 0, Stdout: "output"}
		resp2 := CommandResponse{ExitCode: 0, Stdout: "output"}
		assert.Equal(t, resp1, resp2)
	})
}
