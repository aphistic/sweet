package sweet

import (
	"io"
	"reflect"
	"runtime/pprof"
	"testing"
)

// Since testing.MainStart can change between versions this file exists to abstract
// that part away so we can support multiple versions at once.

func mainStart(s *S) (*testing.M, error) {
	tests := make([]testing.InternalTest, 0)
	tests = append(tests, testing.InternalTest{
		Name: "Tests",
		F: func(t *testing.T) {
			for _, runner := range s.suiteRunners {
				runner.Run(t)
			}
		},
	})
	benchmarks := make([]testing.InternalBenchmark, 0)
	examples := make([]testing.InternalExample, 0)

	fParams := make([]reflect.Value, 0)
	switch {
	case isV1_8MainStart():
		fParams = append(fParams, reflect.ValueOf(testDeps{}))
		fParams = append(fParams, reflect.ValueOf(tests))
		fParams = append(fParams, reflect.ValueOf(benchmarks))
		fParams = append(fParams, reflect.ValueOf(examples))
	case isV1_7MainStart():
		fParams = append(fParams, reflect.ValueOf(func(pat, str string) (bool, error) {
			return false, nil
		}))
		fParams = append(fParams, reflect.ValueOf(tests))
		fParams = append(fParams, reflect.ValueOf(benchmarks))
		fParams = append(fParams, reflect.ValueOf(examples))
	default:
		return nil, errUnsupportedVersion
	}

	fVal := reflect.ValueOf(testing.MainStart)
	res := fVal.Call(fParams)

	if len(res) != 1 {
		return nil, errUnknownResponse
	}

	mIface := res[0]
	if mIface.Kind() != reflect.Ptr {
		return nil, errUnknownResponse
	}

	mVal, ok := mIface.Interface().(*testing.M)
	if !ok {
		return nil, errUnknownResponse
	}

	return mVal, nil
}

// These functions check the testing.MainStart signatures to see which one matches up.  We only go back
// to 1.7 because that's when the subtest functionality that sweet relies on was added. Maybe that's when
// MainStart was added too? I dunno, who cares?

func isV1_7MainStart() bool {
	fType := reflect.TypeOf(testing.MainStart)
	if fType.NumIn() != 4 {
		return false
	}

	// First param is func(string, string) (bool, error)
	param := fType.In(0)
	if param.Kind() != reflect.Func {
		return false
	}
	if param.NumIn() != 2 || param.NumOut() != 2 {
		return false
	}
	if param.In(0).Kind() != reflect.String || param.In(1).Kind() != reflect.String {
		return false
	}
	if param.Out(0).Kind() != reflect.Bool ||
		(param.Out(1).Kind() != reflect.Interface && param.Out(1).Name() != "error") {
		return false
	}

	// Second param is []InternalTest
	param = fType.In(1)
	if param.Kind() != reflect.Slice || param.Elem().Name() != "InternalTest" {
		return false
	}

	// Third param is []InternalBenchmark
	param = fType.In(2)
	if param.Kind() != reflect.Slice || param.Elem().Name() != "InternalBenchmark" {
		return false
	}

	// Fourth param is []InternalExample
	param = fType.In(3)
	if param.Kind() != reflect.Slice || param.Elem().Name() != "InternalExample" {
		return false
	}

	return true
}

func isV1_8MainStart() bool {
	fType := reflect.TypeOf(testing.MainStart)
	if fType.NumIn() != 4 {
		return false
	}

	// First param is an internal interface
	param := fType.In(0)
	if param.Name() != "testDeps" || param.Kind() != reflect.Interface {
		return false
	}

	// Second param is []InternalTest
	param = fType.In(1)
	if param.Kind() != reflect.Slice || param.Elem().Name() != "InternalTest" {
		return false
	}

	// Third param is []InternalBenchmark
	param = fType.In(2)
	if param.Kind() != reflect.Slice || param.Elem().Name() != "InternalBenchmark" {
		return false
	}

	// Fourth param is []InternalExample
	param = fType.In(3)
	if param.Kind() != reflect.Slice || param.Elem().Name() != "InternalExample" {
		return false
	}

	return true
}

// testDeps is an implementation of the testDeps interface required by some
// versions of testing.MainStart
type testDeps struct{}

func (testDeps) ImportPath() string {
	return ""
}

func (testDeps) MatchString(pat, str string) (bool, error) {
	return false, nil
}

func (testDeps) StartCPUProfile(w io.Writer) error {
	return pprof.StartCPUProfile(w)
}

func (testDeps) StopCPUProfile() {
	pprof.StopCPUProfile()
}

// TODO figure out what Go 1.10 is doing with these
func (testDeps) StartTestLog(w io.Writer) {
}

func (testDeps) StopTestLog() error {
	return nil
}

func (testDeps) WriteHeapProfile(w io.Writer) error {
	return pprof.WriteHeapProfile(w)
}

func (testDeps) WriteProfileTo(name string, w io.Writer, debug int) error {
	return pprof.Lookup(name).WriteTo(w, debug)
}
