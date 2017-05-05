package sweet

import "fmt"

type statsPlugin struct {
	suites map[string]*suiteStats
}

type suiteStats struct {
	Name   string
	Passed int
	Failed int
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

}
func (p *statsPlugin) TestStarting(suite, test string) {

}
func (p *statsPlugin) TestPassed(suite, test string, stats *TestPassedStats) {
	s := p.getSuite(suite)
	s.Passed++
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

	if len(sortedNames) > 0 {
		fmt.Printf("Suite Results:\n")
		fmt.Printf("--------------\n")
		for _, name := range sortedNames {
			suite := p.suites[name]
			fmt.Printf("%s - Total: %d, Passed: %d, Failed: %d\n",
				name,
				suite.Passed+suite.Failed,
				suite.Passed,
				suite.Failed)
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
