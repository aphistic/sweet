package sweet

import (
	"fmt"
	"path"
	"reflect"
	"strings"
	"testing"
	"time"
)

type suiteRunner struct {
	s     *S
	suite interface{}

	suiteFailed bool

	suppressDeprecation bool
	deprecatedUsages    []string
}

func newSuiteRunner(s *S, suite interface{}) *suiteRunner {
	return &suiteRunner{
		s:                s,
		suite:            suite,
		deprecatedUsages: make([]string, 0),
	}
}

func (s *suiteRunner) addDeprecationWarning(funcName string) {
	s.deprecatedUsages = append(s.deprecatedUsages, funcName)
}

func (s *suiteRunner) runPlugins(f func(plugin Plugin)) {
	for _, plugin := range s.s.plugins {
		f(plugin)
	}
}

func (s *suiteRunner) Run(t *testing.T) {
	suiteStart := time.Now()

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

	t.Run(suiteName, func(t *testing.T) {
		if *flagParallelSuites {
			t.Parallel()
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

		setUpSuiteVal := suiteVal.MethodByName(defSetUpSuite.Name)
		tearDownSuiteVal := suiteVal.MethodByName(defTearDownSuite.Name)

		v, err := defSetUpSuite.Validate(setUpSuiteVal)
		if err == errDeprecated {
			if !s.suppressDeprecation {
				s.addDeprecationWarning(formatName(suiteName, defSetUpSuite.Name))
			}

			// We handled the error, clear it so the set up still runs
			err = nil
		} else if err == errInvalidValue {
			// Skip running the suite set up because it doesn't exist.
		} else if err != nil {
			panic(fmt.Sprintf("%s has an unsupported method signature",
				formatName(suiteName, defSetUpSuite.Name)))
		}
		if err == nil {
			switch v {
			case 1:
				setUpSuiteVal.Call(nil)
			}
		}

		s.runPlugins(func(plugin Plugin) {
			plugin.SuiteStarting(suiteName)
		})
		for idx := 0; idx < suiteVal.NumMethod(); idx++ {
			methodType := suiteType.Method(idx)

			if methodType.Name[:4] == "Test" {
				methodVal := suiteVal.Method(idx)
				testName := methodType.Name
				t.Run(testName, func(t *testing.T) {
					s.testRunner(
						testName,
						t,
						methodVal,
						suiteName,
						suiteVal,
					)
				})
			}
		}

		v, err = defTearDownSuite.Validate(tearDownSuiteVal)
		if err == errDeprecated {
			if !s.suppressDeprecation {
				s.addDeprecationWarning(formatName(suiteName, defTearDownSuite.Name))
			}

			// We handled the error, clear it so the tear down is still run
			err = nil
		} else if err == errInvalidValue {
			// Continue on because a suite tear down was not provided.
		} else if err != nil {
			panic(fmt.Sprintf("%s has an unsupported method signature",
				formatName(suiteName, defTearDownSuite.Name)))
		}
		if err == nil {
			switch v {
			case 1:
				tearDownSuiteVal.Call(nil)
			}
		}

		s.runPlugins(func(plugin Plugin) {
			plugin.SuiteFinished(suiteName, &SuiteFinishedStats{
				Time: time.Since(suiteStart),
			})
		})
	})
}

func (s *suiteRunner) testRunner(
	testName string,
	t *testing.T,
	methodVal reflect.Value,
	suiteName string,
	suiteVal reflect.Value,
) {
	testFullName := fmt.Sprintf("%s/%s", suiteName, testName)

	setUpTestVal := suiteVal.MethodByName(defSetUpTest.Name)
	tearDownTestVal := suiteVal.MethodByName(defTearDownTest.Name)

	wrapT := newSweetT(t, testName)

	tVal := reflect.ValueOf(t)
	wrapTVal := reflect.ValueOf(wrapT)

	if setUpAllTests != nil {
		setUpAllTests(wrapT)
	}

	v, err := defSetUpTest.Validate(setUpTestVal)
	if err == errDeprecated {
		if !s.suppressDeprecation {
			s.addDeprecationWarning(formatName(suiteName, defSetUpTest.Name))
		}

		// Clean out the error because we've handled it
		err = nil
	} else if err == errInvalidValue {
		// Continue on because a test set up wasn't provided.
	} else if err != nil {
		panic(fmt.Sprintf("%s has an unsupported method signature",
			formatName(suiteName, defSetUpTest.Name)))
	}
	if err == nil {
		switch v {
		case 1:
			setUpTestVal.Call([]reflect.Value{tVal})
		case 2:
			setUpTestVal.Call([]reflect.Value{wrapTVal})
		}
	}

	// Call the actual test function in something that we can recover from
	failureStats := &TestFailedStats{
		Name:   testFullName,
		Frames: make([]*TestFailedFrame, 0),
	}
	testStart := time.Now()
	func() {
		defer func() {
			if r := recover(); r != nil {
				switch result := r.(type) {
				case *testFailed:
					failureStats.Message = result.Message
					failureStats.Frames = make(
						[]*TestFailedFrame,
						len(result.Frames),
					)
					frameCount := len(result.Frames) - 1
					for idx := frameCount; idx >= 0; idx-- {
						frame := result.Frames[idx]
						failureStats.Frames[frameCount-idx] = &TestFailedFrame{
							File: frame.Filename,
							Line: frame.LineNumber,
						}
					}
					wrapT.Fail()
				case *testSkipped:
					// Nothing to do for this because it was handled before
					// the panic
				default:
					panic(r)
				}
			}
		}()

		s.runPlugins(func(plugin Plugin) {
			plugin.TestStarting(suiteName, testName)
		})

		v, err := defTest.Validate(methodVal)
		if err == errDeprecated {
			if !s.suppressDeprecation {
				s.addDeprecationWarning(formatName(suiteName, testName))
			}

			// We handled the error, clear it so the test still runs
			err = nil
		} else if err == errInvalidValue {
			panic("There's a test we can't get info for, contact the Sweet dev!\n")
		} else if err != nil {
			panic(fmt.Sprintf("%s has an unsupported method signature.",
				formatName(suiteName, testName)))
		}
		if err == nil {
			// Run the actual test method
			switch v {
			case 1:
				methodVal.Call([]reflect.Value{tVal})
			case 2:
				methodVal.Call([]reflect.Value{wrapTVal})
			}
		}
	}()

	v, err = defTearDownTest.Validate(tearDownTestVal)
	if err == errDeprecated {
		if !s.suppressDeprecation {
			s.addDeprecationWarning(formatName(suiteName, defTearDownTest.Name))
		}

		// We handled the error, clear it so the clean up still runs
		err = nil
	} else if err == errInvalidValue {
		// Continue on because a test tear down was not provided
	} else if err != nil {
		panic(fmt.Sprintf("%s has an unsupported method signature",
			formatName(suiteName, defTearDownTest.Name)))
	}
	if err == nil {
		switch v {
		case 1:
			tearDownTestVal.Call([]reflect.Value{tVal})
		case 2:
			tearDownTestVal.Call([]reflect.Value{wrapTVal})
		}
	}

	if tearDownAllTests != nil {
		tearDownAllTests(wrapT)
	}

	s.runPlugins(func(plugin Plugin) {
		if wrapT.Failed() {
			plugin.TestFailed(suiteName, testName, failureStats)
		} else if wrapT.Skipped() {
			plugin.TestSkipped(suiteName, testName, &TestSkippedStats{
				Time: time.Since(testStart),
			})
		} else {
			plugin.TestPassed(suiteName, testName, &TestPassedStats{
				Time: time.Since(testStart),
			})
		}
	})

	if wrapT.Failed() {
		s.suiteFailed = true

		fmt.Printf("-------------------------------------------------\n")
		fmt.Printf("FAIL: %s\n\n", failureStats.Name)

		for _, line := range wrapT.output {
			fmt.Print(line)
		}
		if len(wrapT.output) > 0 {
			fmt.Printf("\n\n")
		}

		for _, frame := range failureStats.Frames {
			fmt.Printf("%s:%d\n", path.Base(frame.File), frame.Line)
		}

		fmt.Printf("%s\n\n", failureStats.Message)
	}
}
