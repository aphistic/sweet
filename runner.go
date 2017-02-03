package sweet

import (
	"fmt"
	"reflect"
	"sync"
	"testing"
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

func SetUpAllTests(f func(t *testing.T)) func(t *testing.T) {
	setUpAllTests = f

	return f
}

func RunSuite(t *testing.T, suite interface{}) {
	var testGroup sync.WaitGroup
	runStats := newStats()

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
					setUpTestType, _ := suiteType.MethodByName(setUpTestName)

					if !hasParams(setUpTestType, setUpTestParams) {
						panic(fmt.Sprintf("%s expects %d parameter(s)", setUpTestName, len(setUpTestParams)))
					}

					setUpTestVal.Call([]reflect.Value{tVal})
				}

				// Call the actual test function in something that we can recover from
				func() {
					defer func() {
						if r := recover(); r != nil {
							failure, ok := r.(*testFailure)
							if !ok {
								panic(r)
							}

							fmt.Printf("-------------------------------------------------\n")
							fmt.Printf("FAIL: %s\n\n%s:%d\n",
								testName,
								failure.Filename, failure.LineNumber)
							fmt.Printf("%s\n\n", failure.Message)

							runStats.Failed++
						} else {
							runStats.Passed++
						}

						testGroup.Done()
					}()

					testGroup.Add(1)
					methodVal.Call([]reflect.Value{tVal})
				}()

				if tearDownTestVal.IsValid() && tearDownTestVal.Kind() == reflect.Func {
					tearDownTestType, _ := suiteType.MethodByName(tearDownTestName)

					if !hasParams(tearDownTestType, tearDownTestParams) {
						panic(fmt.Sprintf("%s expects %d parameter(s)", tearDownTestName, len(tearDownTestParams)))
					}

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

	testGroup.Wait()
	fmt.Printf("%s - Total: %d, Passed: %d, Failed: %d\n",
		suiteName,
		runStats.Passed+runStats.Failed,
		runStats.Passed,
		runStats.Failed)
}
