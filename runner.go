package sweet

import (
	"flag"
	"fmt"
	"os"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"testing"
)

var (
	setUpAllTests    func(t T)
	tearDownAllTests func(t T)
)

func hasParams(method reflect.Method, paramTypes []reflect.Type) bool {
	if method.Type.NumIn() != len(paramTypes)+1 {
		return false
	}
	return true
}

type registeredOptions struct {
	Plugin  Plugin
	Options *PluginOptions
}

type S struct {
	suiteRunners     []*suiteRunner
	deprecatedSuites map[interface{}]bool

	plugins []Plugin
	options map[string]*registeredOptions
}

func Run(m *testing.M, f func(s *S)) {
	if !flag.Parsed() {
		flag.Parse()
	}

	s := &S{
		suiteRunners:     make([]*suiteRunner, 0),
		deprecatedSuites: make(map[interface{}]bool),

		plugins: make([]Plugin, 0),
		options: make(map[string]*registeredOptions),
	}
	s.RegisterPlugin(newStatsPlugin())

	f(s)

	if *flagHelp {
		fmt.Println("Sweet Options")
		fmt.Println("=============")

		fmt.Println("-sweet.help: Displays this help text")
		fmt.Println("-sweet.opt: Passes an argument to a sweet plugin.")
		fmt.Println("            Ex: -sweet.opt \"plug.myopt=myval\"")
		fmt.Println("-sweet.include: Only run tests that match the provided expression")
		fmt.Println("-sweet.exclude: Do not include tests that match the provided expression")
		fmt.Println("-sweet.extended: Show extended error information for failed tests")
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

	newM, err := mainStart(s)
	if err == errUnsupportedVersion {
		fmt.Fprintf(os.Stderr,
			"This version of Go is unsupported by Sweet. Please open an issue mentioning the\n"+
				"version \"%s\" at the project page: https://www.github.com/aphistic/sweet/\n",
			runtime.Version())
		os.Exit(1)
	} else if err != nil {
		fmt.Fprintf(os.Stderr,
			"Error while setting up tests: %s\n", err)
		os.Exit(1)
	}

	code := newM.Run()

	for _, plugin := range s.plugins {
		plugin.Finished()
	}

	const deprecationExampleLength = 3
	deprecatedUsages := make([]string, 0)
	for _, runner := range s.suiteRunners {
		uniqueNames := make(map[string]bool)
		for _, dep := range runner.deprecatedUsages {
			uniqueNames[dep] = true
		}
		for dep := range uniqueNames {
			deprecatedUsages = append(deprecatedUsages, dep)
		}
	}
	sort.Strings(deprecatedUsages)

	if len(deprecatedUsages) > 0 {
		fmt.Fprintf(os.Stderr,
			"Some test methods are using a deprecated version of the method signature. "+
				"Please update them to the latest version as seen in the documentation at "+
				"https://github.com/aphistic/sweet. Some of the deprecated usages are the "+
				"following:\n",
		)

		for idx, usage := range deprecatedUsages {
			additionalUsages := len(deprecatedUsages) - deprecationExampleLength

			fmt.Fprintf(os.Stderr, usage)
			if idx < deprecationExampleLength-1 {
				fmt.Fprintf(os.Stderr, ",")
			} else if idx >= deprecationExampleLength-1 {
				if additionalUsages > 0 {
					fmt.Fprintf(os.Stderr, "... %d other(s).\n", additionalUsages)
				} else {
					fmt.Fprintf(os.Stderr, "\n")
				}
				break
			}

			fmt.Fprintf(os.Stderr, " ")
		}

		fmt.Fprintf(os.Stderr, "\n")
	}

	os.Exit(code)
}

func (s *S) SetUpAllTests(f func(t T)) func(t T) {
	setUpAllTests = f

	return f
}
func (s *S) TearDownAllTests(f func(t T)) func(t T) {
	tearDownAllTests = f

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

func (s *S) AddSuite(suite interface{}) {
	s.suiteRunners = append(s.suiteRunners, newSuiteRunner(s, suite))
}

func (s *S) suppressDeprecation(suite interface{}) {
	for _, runner := range s.suiteRunners {
		if runner.suite == suite {
			runner.suppressDeprecation = true
		}
	}
}
