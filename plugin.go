package sweet

import (
	"time"
)

type Plugin interface {
	Name() string
	Options() *PluginOptions
	SetOption(name, value string)

	Starting()
	SuiteStarting(suite string)
	TestStarting(suite, test string)
	TestPassed(suite, test string, stats *TestPassedStats)
	TestFailed(suite, test string, stats *TestFailedStats)
	TestSkipped(suite, test string, stats *TestSkippedStats)
	SuiteFinished(suite string, stats *SuiteFinishedStats)
	Finished()
}

type PluginOptions struct {
	Prefix  string
	Options map[string]*PluginOption
}
type PluginOption struct {
	Help    string
	Default string
}

type TestPassedStats struct {
	Time time.Duration
}

type TestFailedStats struct {
	Name    string
	Time    time.Duration
	Message string
	Frames  []*TestFailedFrame
}
type TestFailedFrame struct {
	File string
	Line int
}

type TestSkippedStats struct {
	Time time.Duration
}

type SuiteFinishedStats struct {
	Time time.Duration
}
