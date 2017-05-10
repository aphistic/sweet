package sweet

import (
	"fmt"
	"os"
	"path"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"
)

type suiteRunner struct {
	s     *S
	suite interface{}
}

func newSuiteRunner(s *S, suite interface{}) *suiteRunner {
	return &suiteRunner{
		s:     s,
		suite: suite,
	}
}

func (s *suiteRunner) runPlugins(f func(plugin Plugin)) {
	for _, plugin := range s.s.plugins {
		f(plugin)
	}
}

func (s *suiteRunner) Run(t *testing.T) {
	suiteStart := time.Now()

	var testGroup sync.WaitGroup

	suiteVal := reflect.ValueOf(s.suite)
	suiteType := suiteVal.Type()

	suiteName := suiteType.Name()
	if suiteVal.CanInterface() {
		suiteIfaceVal := reflect.Indirect(suiteVal)
		if suiteIfaceVal == reflect.Zero(suiteType) {
			suiteName = "UnknownSuite"
		} else {
			suiteName = suiteIfaceVal.Type().Name()
		}
	}

	lowerName := strings.ToLower(suiteName)
	if len(flagInclude) > 0 {
		found := false
		for _, name := range flagInclude {
			if strings.ToLower(name) == lowerName {
				found = true
				break
			}
		}
		if !found {
			return
		}
	}
	if len(flagExclude) > 0 {
		for _, name := range flagExclude {
			if strings.ToLower(name) == lowerName {
				return
			}
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
		plugin.SuiteStarting(suiteName)
	})
	for idx := 0; idx < suiteVal.NumMethod(); idx++ {
		methodType := suiteType.Method(idx)

		if methodType.Name[:4] == "Test" {
			methodVal := suiteVal.Method(idx)
			testName := methodType.Name
			testFullName := fmt.Sprintf("%s/%s", suiteName, methodType.Name)
			testStart := time.Now()
			t.Run(testFullName, func(t *testing.T) {
				// Start capturing stdout so we only display it when a test fails.
				oldStdout := os.Stdout
				stdout, err := newPipeCapture()
				if err != nil {
					t.Errorf("Unable to create IO pipe: %s", err)
					return
				}
				os.Stdout = stdout.W()

				if setUpAllTests != nil {
					setUpAllTests(t)
				}

				tVal := reflect.ValueOf(t)
				if setUpTestVal.IsValid() && setUpTestVal.Kind() == reflect.Func {
					setUpTestType, _ := suiteType.MethodByName(setUpTestName)

					if !hasParams(setUpTestType, setUpTestParams) {
						t.Errorf("%s expects %d parameter(s)", setUpTestName, len(setUpTestParams))
						return
					}

					setUpTestVal.Call([]reflect.Value{tVal})
				}

				// Call the actual test function in something that we can recover from
				failureStats := &TestFailedStats{
					Frames: make([]*TestFailedFrame, 0),
				}
				func() {
					defer func() {
						if r := recover(); r != nil {
							failure, ok := r.(*testFailure)
							if !ok {
								panic(r)
							}

							failureStats.Name = testFullName
							failureStats.Message = failure.Message
							failureStats.Frames = make([]*TestFailedFrame, len(failure.Frames))
							frameCount := len(failure.Frames) - 1
							for idx := frameCount; idx >= 0; idx-- {
								frame := failure.Frames[idx]
								failureStats.Frames[frameCount-idx] = &TestFailedFrame{
									File: frame.Filename,
									Line: frame.LineNumber,
								}
							}

							t.Fail()
						}

						testGroup.Done()
					}()

					testGroup.Add(1)
					s.runPlugins(func(plugin Plugin) {
						plugin.TestStarting(suiteName, testName)
					})
					methodVal.Call([]reflect.Value{tVal})
				}()

				if tearDownTestVal.IsValid() && tearDownTestVal.Kind() == reflect.Func {
					tearDownTestType, _ := suiteType.MethodByName(tearDownTestName)

					if !hasParams(tearDownTestType, tearDownTestParams) {
						t.Errorf("%s expects %d parameter(s)", tearDownTestName, len(tearDownTestParams))
						return
					}

					tearDownTestVal.Call([]reflect.Value{tVal})
				}

				if tearDownAllTests != nil {
					tearDownAllTests(t)
				}

				s.runPlugins(func(plugin Plugin) {
					if t.Failed() {
						plugin.TestFailed(suiteName, testName, failureStats)
					} else {
						plugin.TestPassed(suiteName, testName, &TestPassedStats{
							Time: time.Since(testStart),
						})
					}
				})

				// Restore stdout before we try printing to it
				os.Stdout = oldStdout
				stdout.Close()

				if t.Failed() {
					fmt.Printf("-------------------------------------------------\n")
					fmt.Printf("FAIL: %s\n\n", failureStats.Name)

					for _, frame := range failureStats.Frames {
						fmt.Printf("%s:%d\n", path.Base(frame.File), frame.Line)
					}
					fmt.Printf("%s\n\n", failureStats.Message)

					if len(stdout.Buffer()) > 0 {
						fmt.Println("stdout:")
						fmt.Print(string(stdout.Buffer()))
						fmt.Println()
					}
				}
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
		plugin.SuiteFinished(suiteName, &SuiteFinishedStats{
			Time: time.Since(suiteStart),
		})
	})
}
