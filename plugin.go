package sweet

import (
	"time"
)

type Plugin interface {
	Name() string

	Starting()
	SuiteStarting(suite string)
	TestStarting(suite, test string)
	TestPassed(suite, test string, stats *TestPassedStats)
	TestFailed(suite, test string, stats *TestFailedStats)
	SuiteFinished(suite string, stats *SuiteFinishedStats)
	Finished()
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
