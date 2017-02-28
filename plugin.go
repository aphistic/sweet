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
	Time    time.Duration
	File    string
	Line    int
	Message string
}

type SuiteFinishedStats struct {
	Time time.Duration
}
