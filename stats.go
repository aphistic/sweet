package sweet

import "fmt"

import "sort"

type statsPlugin struct {
	suites map[string]*suiteStats
}

type suiteStats struct {
	Name    string
	Passed  int
	Failed  int
	Skipped int
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
	s.Passed++
}
func (p *statsPlugin) TestSkipped(suite, test string, stats *TestSkippedStats) {
	s := p.getSuite(suite)
	s.Skipped++
}
func (p *statsPlugin) TestFailed(suite, test string, stats *TestFailedStats) {
	s := p.getSuite(suite)
	s.Failed++
}
func (p *statsPlugin) SuiteFinished(suite string, stats *SuiteFinishedStats) {

}
func (p *statsPlugin) Finished() {
	sortedNames := make([]string, 0)
	for key := range p.suites {
		sortedNames = append(sortedNames, key)
	}
	sort.Strings(sortedNames)

	if len(sortedNames) > 0 {
		fmt.Println("")
		fmt.Printf("Suite Results:\n")
		fmt.Printf("--------------\n")
		for _, name := range sortedNames {
			suite := p.suites[name]
			fmt.Printf("%s - Total: %d, Passed: %d, Failed: %d, Skipped: %d\n",
				name,
				suite.Passed+suite.Failed+suite.Skipped,
				suite.Passed,
				suite.Failed,
				suite.Skipped)
		}
		fmt.Println("")
	}
}

func (p *statsPlugin) getSuite(name string) *suiteStats {
	suite, ok := p.suites[name]
	if !ok {
		suite = &suiteStats{
			Name: name,
		}
		p.suites[name] = suite
	}
	return suite
}
