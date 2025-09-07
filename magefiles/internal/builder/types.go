package builder

import "errors"

type Config struct {
	MainBinary string
	MCPBinary  string
	BinDir     string
	GOPATHBin  string
}

type BuildOptions struct {
	TrimPath bool
	LDFlags  string
}

var ErrSourceBinaryNotExist = errors.New("source binary does not exist")
