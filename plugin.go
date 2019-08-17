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
	TestStarting(testName *TestName)
	TestPassed(testName *TestName, stats *TestPassedStats)
	TestFailed(testName *TestName, stats *TestFailedStats)
	TestSkipped(testName *TestName, stats *TestSkippedStats)
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
	Name    *TestName
	Time    time.Duration
	Message string
	Frames  []*TestFailedFrame
}
type TestFailedFrame struct {
	File   string
	Line   int
	Hidden bool
}

type TestSkippedStats struct {
	Time time.Duration
}

type SuiteFinishedStats struct {
	Time time.Duration
}
