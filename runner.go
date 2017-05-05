package sweet

import (
	"fmt"
	"os"
	"path"
	"reflect"
	"sort"
	"strings"
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
	Name  string
	Suite interface{}
}

type registeredOptions struct {
	Plugin  Plugin
	Options *PluginOptions
}

type S struct {
	suites  []*suiteRunner
	plugins []Plugin
	options map[string]*registeredOptions
}

func T(f func(s *S)) {
	if *flagHelp {
		flagSkipRuns = true
	}

	s := &S{
		suites:  make([]*suiteRunner, 0),
		plugins: make([]Plugin, 0),
		options: make(map[string]*registeredOptions),
	}
	s.RegisterPlugin(newStatsPlugin())

	f(s)

	if !flagSkipRuns {
		s.runPlugins(func(plugin Plugin) {
			plugin.Finished()
		})
	}

	if *flagHelp {
		fmt.Println("Sweet Options")
		fmt.Println("=============")

		fmt.Println("-sweet.help: Displays this help text")
		fmt.Println("-sweet.opt: Passes an argument to a sweet plugin.")
		fmt.Println("            Ex: -sweet.opt \"plug.myopt=myval\"")
		fmt.Println("")

		sortedPrefixes := make([]string, 0)
		for prefix := range s.options {
			sortedPrefixes = append(sortedPrefixes, prefix)
		}
		sort.Strings(sortedPrefixes)

		for _, prefix := range sortedPrefixes {
			opts := s.options[prefix]

			for optionName, optSetting := range opts.Options.Options {
				fmt.Printf("  %s.%s - %s\n", prefix, optionName, optSetting.Help)
			}
		}

		fmt.Println("")

		os.Exit(0)
	}
}

func (s *S) SetUpAllTests(f func(t *testing.T)) func(t *testing.T) {
	setUpAllTests = f

	return f
}

func (s *S) RegisterPlugin(plugin Plugin) {
	if plugin == nil {
		return
	}

	plugOpts := plugin.Options()
	if plugOpts != nil {
		if oldOpts, ok := s.options[plugOpts.Prefix]; ok {
			fmt.Fprintf(os.Stderr,
				"ERROR: Sweet option prefix \"%s\" has already been registered by %s.\n",
				oldOpts.Options.Prefix, oldOpts.Plugin.Name())
			fmt.Fprintf(os.Stderr,
				"You may have accidentally registered the same plugin twice or a plugin you're\n"+
					"using may have a colliding options prefix with another you're using.\n\n")
			os.Exit(1)
		}

		s.options[plugOpts.Prefix] = &registeredOptions{
			Plugin:  plugin,
			Options: plugOpts,
		}

		// When registering a plugin, go through each option and see if it matches
		// a plugin option.  If it does, set the option in the plugin

		for _, opt := range flagOpts {
			valIdx := strings.Index(opt, "=")
			value := ""
			if valIdx >= 0 {
				value = opt[valIdx+1:]
			}
			name := opt[:valIdx]

			if strings.HasPrefix(name, plugOpts.Prefix+".") {
				plugName := name[len(plugOpts.Prefix+"."):]
				plugin.SetOption(plugName, value)
			}
		}
	}

	s.plugins = append(s.plugins, plugin)
}

func (s *S) runPlugins(f func(plugin Plugin)) {
	for _, plugin := range s.plugins {
		f(plugin)
	}
}

func (s *S) RunSuite(t *testing.T, suite interface{}) {
	if flagSkipRuns {
		return
	}

	suiteStart := time.Now()

	runner := &suiteRunner{
		Suite: suite,
	}
	s.suites = append(s.suites, runner)

	var testGroup sync.WaitGroup

	suiteVal := reflect.ValueOf(runner.Suite)
	suiteType := suiteVal.Type()

	runner.Name = suiteType.Name()
	if suiteVal.CanInterface() {
		suiteIfaceVal := reflect.Indirect(suiteVal)
		if suiteIfaceVal == reflect.Zero(suiteType) {
			runner.Name = "UnknownSuite"
		} else {
			runner.Name = suiteIfaceVal.Type().Name()
		}
	}

	lowerName := strings.ToLower(runner.Name)
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
		plugin.SuiteStarting(runner.Name)
	})
	for idx := 0; idx < suiteVal.NumMethod(); idx++ {
		methodType := suiteType.Method(idx)

		if methodType.Name[:4] == "Test" {
			methodVal := suiteVal.Method(idx)
			testName := methodType.Name
			testFullName := fmt.Sprintf("%s/%s", runner.Name, methodType.Name)
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

							failureStats.Message = failure.Message

							fmt.Printf("-------------------------------------------------\n")
							fmt.Printf("FAIL: %s\n\n", testFullName)

							failureStats.Frames = make([]*TestFailedFrame, len(failure.Frames))
							frameCount := len(failure.Frames) - 1
							for idx := frameCount; idx >= 0; idx-- {
								frame := failure.Frames[idx]
								failureStats.Frames[frameCount-idx] = &TestFailedFrame{
									File: frame.Filename,
									Line: frame.LineNumber,
								}
								fmt.Printf("%s:%d\n", path.Base(frame.Filename), frame.LineNumber)
							}
							fmt.Printf("%s\n\n", failure.Message)

							t.Fail()
						}

						testGroup.Done()
					}()

					testGroup.Add(1)
					s.runPlugins(func(plugin Plugin) {
						plugin.TestStarting(runner.Name, testName)
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
					if t.Failed() {
						plugin.TestFailed(runner.Name, testName, failureStats)
					} else {
						plugin.TestPassed(runner.Name, testName, &TestPassedStats{
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
		plugin.SuiteFinished(runner.Name, &SuiteFinishedStats{
			Time: time.Since(suiteStart),
		})
	})
}
