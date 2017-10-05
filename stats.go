package sweet

import (
	"fmt"
	"os"
	"sort"
	"sync"
	"sync/atomic"

	"github.com/mgutz/ansi"
	"golang.org/x/crypto/ssh/terminal"
)

type statsPlugin struct {
	suitesLock sync.Mutex
	suites     map[string]*suiteStats
}

type suiteStats struct {
	Name    string
	Passed  int64
	Failed  int64
	Skipped int64
}

func newStatsPlugin() *statsPlugin {
	return &statsPlugin{
		suites: make(map[string]*suiteStats),
	}
}

func (p *statsPlugin) Name() string {
	return "Test Stats"
}

func (p *statsPlugin) Options() *PluginOptions {
	return nil
}

func (p *statsPlugin) SetOption(name, value string) {

}

func (p *statsPlugin) Starting() {

}
func (p *statsPlugin) SuiteStarting(suite string) {
	// Get the suite so stats are aware of it and it shows up
	// in the final results
	p.getSuite(suite)
}
func (p *statsPlugin) TestStarting(suite, test string) {

}
func (p *statsPlugin) TestPassed(suite, test string, stats *TestPassedStats) {
	s := p.getSuite(suite)
	atomic.AddInt64(&s.Passed, 1)
}
func (p *statsPlugin) TestSkipped(suite, test string, stats *TestSkippedStats) {
	s := p.getSuite(suite)
	atomic.AddInt64(&s.Skipped, 1)
}
func (p *statsPlugin) TestFailed(suite, test string, stats *TestFailedStats) {
	s := p.getSuite(suite)
	atomic.AddInt64(&s.Failed, 1)
}
func (p *statsPlugin) SuiteFinished(suite string, stats *SuiteFinishedStats) {

}
func (p *statsPlugin) Finished() {
	sortedNames := make([]string, 0)
	for key := range p.suites {
		sortedNames = append(sortedNames, key)
	}
	sort.Strings(sortedNames)

	out := os.Stdout
	isTerm := terminal.IsTerminal(int(out.Fd()))

	passColor := ansi.ColorFunc("green")
	failColor := ansi.ColorFunc("red")
	skipColor := ansi.ColorFunc("yellow")

	if len(sortedNames) > 0 {
		fmt.Fprintln(out, "")
		fmt.Fprintf(out, "Suite Results:\n")
		fmt.Fprintf(out, "--------------\n")
		for _, name := range sortedNames {
			suite := p.suites[name]

			totalStr := fmt.Sprintf("%d", suite.Passed+suite.Failed+suite.Skipped)

			passedStr := fmt.Sprintf("%d", suite.Passed)
			if isTerm && suite.Passed > 0 {
				passedStr = passColor(passedStr)
			}

			failedStr := fmt.Sprintf("%d", suite.Failed)
			if isTerm && suite.Failed > 0 {
				failedStr = failColor(failedStr)
			}

			skippedStr := fmt.Sprintf("%d", suite.Skipped)
			if isTerm && suite.Skipped > 0 {
				skippedStr = skipColor(skippedStr)
			}

			fmt.Fprintf(out, "%s - Total: %s, Passed: %s, Failed: %s, Skipped: %s\n",
				name,
				totalStr,
				passedStr,
				failedStr,
				skippedStr,
			)
		}
		fmt.Fprintln(out, "")
	}
}

func (p *statsPlugin) getSuite(name string) *suiteStats {
	p.suitesLock.Lock()
	defer p.suitesLock.Unlock()
	suite, ok := p.suites[name]
	if !ok {
		suite = &suiteStats{
			Name: name,
		}
		p.suites[name] = suite
	}
	return suite
}
