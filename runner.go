package sweet

import (
	"fmt"
	"os"
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

	if *flagInclude != "" && *flagInclude != runner.Name {
		return
	}
	if *flagExclude != "" && *flagExclude == runner.Name {
		return
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
					if testFailed {
						plugin.TestFailed(runner.Name, testName, failureStats)
					} else {
						plugin.TestPassed(runner.Name, testName, &TestPassedStats{
							Time: time.Since(testStart),
						})
					}
				})

				if testFailed {
					t.Fail()
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
		plugin.SuiteFinished(runner.Name, &SuiteFinishedStats{
			Time: time.Since(suiteStart),
		})
	})
}
