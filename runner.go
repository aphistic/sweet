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
	setUpAllTests    func(t *testing.T)
	tearDownAllTests func(t *testing.T)

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

type registeredOptions struct {
	Plugin  Plugin
	Options *PluginOptions
}

type S struct {
	addedSuites []interface{}

	plugins []Plugin
	options map[string]*registeredOptions
}

func Run(m *testing.M, f func(s *S)) {
	if !flag.Parsed() {
		flag.Parse()
	}

	s := &S{
		addedSuites: make([]interface{}, 0),
		plugins:     make([]Plugin, 0),
		options:     make(map[string]*registeredOptions),
	}
	s.RegisterPlugin(newStatsPlugin())

	f(s)

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

	os.Exit(code)
}

func (s *S) SetUpAllTests(f func(t *testing.T)) func(t *testing.T) {
	setUpAllTests = f

	return f
}
func (s *S) TearDownAllTests(f func(t *testing.T)) func(t *testing.T) {
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
	s.addedSuites = append(s.addedSuites, suite)
}
