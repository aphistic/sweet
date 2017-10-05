package sweet

import (
	"os"
	"runtime"
	"strings"
)

var (
	goPackages = []string{
		"runtime",
		"reflect",
	}
)

type testCompletion interface{}

type failureFrame struct {
	Filename   string
	LineNumber int
}

type testFailed struct {
	Message string
	Frames  []*failureFrame
}

type testSkipped struct{}

func isGoPackage(path string) bool {
	srcDelim := "src" + string(os.PathSeparator)
	idx := strings.Index(path, srcDelim)
	if idx < 0 {
		return false
	}
	idx += len(srcDelim)
	for _, pkg := range goPackages {
		if strings.HasPrefix(path[idx:], pkg) {
			return true
		}
	}

	return false
}

func skipTest(message string) {
	skipped := &testSkipped{}
	panic(skipped)
}

func failTest(message string, callerSkip ...int) {
	failFrames := make([]*failureFrame, 0)
	if len(callerSkip) > 0 {
		callIdx := callerSkip[0] + 2
		callers := make([]uintptr, 0)
		for {
			pc, file, _, ok := runtime.Caller(callIdx)
			if ok {
				if isGoPackage(file) {
					break
				}
				callers = append(callers, pc)
				callIdx++
			} else {
				break
			}
		}

		frames := runtime.CallersFrames(callers)
		for {
			frame, more := frames.Next()
			failFrames = append(failFrames, &failureFrame{
				Filename:   frame.File,
				LineNumber: frame.Line,
			})

			if !more {
				break
			}
		}
	}

	failure := &testFailed{
		Message: message,
		Frames:  failFrames,
	}

	panic(failure)
}

// GomegaFail is a utility function provided to hook into the Gomega matcher library. To use
// this it's easiest to do the following in your set up:
//   func TestMain(m *testing.M) {
//       RegisterFailHandler(sweet.GomegaFail)
//
//       sweet.Run(m, func(s *sweet.S) {
//           // ... Suite set up ...
//       })
//   }
func GomegaFail(message string, callerSkip ...int) {
	failTest(message, callerSkip...)
}
