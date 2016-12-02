package sweet

import (
	"fmt"
	"reflect"
	"testing"
)

const (
	setUpTestName    = "SetUpTest"
	tearDownTestName = "TearDownTest"
)

var (
	setUpAllTests func(t *testing.T) = nil

	setUpSuiteName   = "SetUpSuite"
	setUpSuiteParams = []reflect.Type{}

	tearDownSuiteName   = "TearDownSuite"
	tearDownSuiteParams = []reflect.Type{}
)

func hasParams(method reflect.Method, paramTypes []reflect.Type) bool {
	if method.Type.NumIn() != len(paramTypes)+1 {
		return false
	}
	return true
}

func SetUpAllTests(f func(t *testing.T)) func(t *testing.T) {
	setUpAllTests = f

	return f
}

func RunSuite(t *testing.T, suite interface{}) {
	suiteVal := reflect.ValueOf(suite)
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

	for idx := 0; idx < suiteVal.NumMethod(); idx++ {
		methodType := suiteType.Method(idx)

		if methodType.Name[:4] == "Test" {
			methodVal := suiteVal.Method(idx)
			testName := fmt.Sprintf("%s/%s", suiteName, methodType.Name)
			t.Run(testName, func(t *testing.T) {
				if setUpAllTests != nil {
					setUpAllTests(t)
				}

				tVal := reflect.ValueOf(t)
				if setUpTestVal.IsValid() && setUpTestVal.Kind() == reflect.Func {
					setUpTestVal.Call([]reflect.Value{tVal})
				}
				methodVal.Call([]reflect.Value{tVal})
				if tearDownTestVal.IsValid() && tearDownTestVal.Kind() == reflect.Func {
					tearDownTestVal.Call([]reflect.Value{tVal})
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
}
