package sweet

import (
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"
)

var (
	setUpAllTests func(t *testing.T)

	setUpSuiteName   = "SetUpSuite"
	setUpSuiteParams = []reflect.Type{}

	setUpTestName   = "SetUpTest"
	setUpTestParams = []reflect.Type{
		reflect.TypeOf(&testing.T{}),
	}

	tearDownTestName   = "TearDownTest"
	tearDownTestParams = []reflect.Type{
		reflect.TypeOf(&testing.T{}),
	}

	tearDownSuiteName   = "TearDownSuite"
	tearDownSuiteParams = []reflect.Type{}
)

func hasParams(method reflect.Method, paramTypes []reflect.Type) bool {
	if method.Type.NumIn() != len(paramTypes)+1 {
		return false
	}
	return true
}

type suiteRunner struct {
	name  string
	suite interface{}
}

type S struct {
	suites  []*suiteRunner
	plugins []Plugin
}

func T(f func(s *S)) {
	s := &S{
		suites:  make([]*suiteRunner, 0),
		plugins: make([]Plugin, 0),
	}
	s.RegisterPlugin(newStatsPlugin())

	f(s)

	s.runPlugins(func(plugin Plugin) {
		plugin.Finished()
	})
}

func (s *S) SetUpAllTests(f func(t *testing.T)) func(t *testing.T) {
	setUpAllTests = f

	return f
}

func (s *S) RegisterPlugin(plugin Plugin) {
	s.plugins = append(s.plugins, plugin)
}

func (s *S) runPlugins(f func(plugin Plugin)) {
	for _, plugin := range s.plugins {
		f(plugin)
	}
}

func (s *S) RunSuite(t *testing.T, suite interface{}) {
	suiteStart := time.Now()

	runner := &suiteRunner{
		suite: suite,
	}
	s.suites = append(s.suites, runner)

	var testGroup sync.WaitGroup

	suiteVal := reflect.ValueOf(runner.suite)
	suiteType := suiteVal.Type()

	runner.name = suiteType.Name()
	if suiteVal.CanInterface() {
		suiteIfaceVal := reflect.Indirect(suiteVal)
		if suiteIfaceVal == reflect.Zero(suiteType) {
			runner.name = "UnknownSuite"
		} else {
			runner.name = suiteIfaceVal.Type().Name()
		}
	}

	setUpSuiteVal := suiteVal.MethodByName(setUpSuiteName)
	tearDownSuiteVal := suiteVal.MethodByName(tearDownSuiteName)

	setUpTestVal := suiteVal.MethodByName(setUpTestName)
	tearDownTestVal := suiteVal.MethodByName(tearDownTestName)

	if setUpSuiteVal.IsValid() && setUpSuiteVal.Kind() == reflect.Func {
		setUpSuiteType, _ := suiteType.MethodByName(setUpSuiteName)

		if !hasParams(setUpSuiteType, setUpSuiteParams) {
			panic(fmt.Sprintf("%s expects %d parameter(s)", setUpSuiteName, len(setUpSuiteParams)))
		}

		setUpSuiteVal.Call(nil)
	}

	s.runPlugins(func(plugin Plugin) {
		plugin.SuiteStarting(runner.name)
	})
	for idx := 0; idx < suiteVal.NumMethod(); idx++ {
		methodType := suiteType.Method(idx)

		if methodType.Name[:4] == "Test" {
			methodVal := suiteVal.Method(idx)
			testName := methodType.Name
			testFullName := fmt.Sprintf("%s/%s", runner.name, methodType.Name)
			testStart := time.Now()
			t.Run(testFullName, func(t *testing.T) {
				if setUpAllTests != nil {
					setUpAllTests(t)
				}

				tVal := reflect.ValueOf(t)
				if setUpTestVal.IsValid() && setUpTestVal.Kind() == reflect.Func {
					setUpTestType, _ := suiteType.MethodByName(setUpTestName)

					if !hasParams(setUpTestType, setUpTestParams) {
						panic(fmt.Sprintf("%s expects %d parameter(s)", setUpTestName, len(setUpTestParams)))
					}

					setUpTestVal.Call([]reflect.Value{tVal})
				}

				// Call the actual test function in something that we can recover from
				testFailed := false
				failureStats := &TestFailedStats{}
				func() {
					defer func() {
						if r := recover(); r != nil {
							failure, ok := r.(*testFailure)
							if !ok {
								panic(r)
							}

							failureStats.File = failure.Filename
							failureStats.Line = failure.LineNumber
							failureStats.Message = failure.Message

							fmt.Printf("-------------------------------------------------\n")
							fmt.Printf("FAIL: %s\n\n%s:%d\n",
								testFullName,
								failure.Filename, failure.LineNumber)
							fmt.Printf("%s\n\n", failure.Message)

							testFailed = true
						}

						testGroup.Done()
					}()

					testGroup.Add(1)
					s.runPlugins(func(plugin Plugin) {
						plugin.TestStarting(runner.name, testName)
					})
					methodVal.Call([]reflect.Value{tVal})
				}()

				if tearDownTestVal.IsValid() && tearDownTestVal.Kind() == reflect.Func {
					tearDownTestType, _ := suiteType.MethodByName(tearDownTestName)

					if !hasParams(tearDownTestType, tearDownTestParams) {
						panic(fmt.Sprintf("%s expects %d parameter(s)", tearDownTestName, len(tearDownTestParams)))
					}

					tearDownTestVal.Call([]reflect.Value{tVal})
				}

				s.runPlugins(func(plugin Plugin) {
					if testFailed {
						plugin.TestFailed(runner.name, testName, failureStats)
					} else {
						plugin.TestPassed(runner.name, testName, &TestPassedStats{
							Time: time.Since(testStart),
						})
					}
				})
			})
		}
	}

	if tearDownSuiteVal.IsValid() && tearDownSuiteVal.Kind() == reflect.Func {
		tearDownSuiteType, _ := suiteType.MethodByName(tearDownSuiteName)

		if !hasParams(tearDownSuiteType, tearDownSuiteParams) {
			panic(fmt.Sprintf("%s expects %d parameter(s)", tearDownSuiteName, len(tearDownSuiteParams)))
		}

		tearDownSuiteVal.Call(nil)
	}

	testGroup.Wait()

	s.runPlugins(func(plugin Plugin) {
		plugin.SuiteFinished(runner.name, &SuiteFinishedStats{
			Time: time.Since(suiteStart),
		})
	})
}
