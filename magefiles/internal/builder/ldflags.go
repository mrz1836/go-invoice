package builder

import (
	"time"
)

type LDFlagsBuilder struct {
	shell ShellRunner
}

func NewLDFlagsBuilder(shell ShellRunner) *LDFlagsBuilder {
	return &LDFlagsBuilder{shell: shell}
}

func (l *LDFlagsBuilder) Build(pkg string) string {
	version := l.getVersion()
	commit := l.getCommit()
	buildDate := time.Now().UTC().Format(time.RFC3339)

	return "-s -w -X " + pkg + ".version=" + version + " -X " + pkg + ".commit=" + commit + " -X " + pkg + ".buildDate=" + buildDate
}

func (l *LDFlagsBuilder) BuildWithVersion(pkg, version string) string {
	commit := l.getCommit()
	buildDate := time.Now().UTC().Format(time.RFC3339)

	return "-s -w -X " + pkg + ".version=" + version + " -X " + pkg + ".commit=" + commit + " -X " + pkg + ".buildDate=" + buildDate
}

func (l *LDFlagsBuilder) getVersion() string {
	version, err := l.shell.Output("git", "describe", "--tags", "--always", "--dirty")
	if err != nil {
		return "dev"
	}
	return version
}

func (l *LDFlagsBuilder) getCommit() string {
	commit, err := l.shell.Output("git", "rev-parse", "--short", "HEAD")
	if err != nil {
		return "unknown"
	}
	return commit
}
