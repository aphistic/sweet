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

type failureFrame struct {
	Filename   string
	LineNumber int
}

type testFailure struct {
	Message string
	Frames  []*failureFrame
}

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

// GomegaFail is a utility function provided to hook into the Gomega matcher library. To use
// this it's easiest to do the following in your set up test:
//   func Test(t *testing.T) {
//       RegisterFailHandler(sweet.GomegaFail)
//
//       sweet.T(func(s *sweet.S) {
//           // ... Suite set up ...
//       })
//   }
func GomegaFail(message string, callerSkip ...int) {
	failFrames := make([]*failureFrame, 0)
	if len(callerSkip) > 0 {
		callIdx := callerSkip[0] + 1
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

	failure := &testFailure{
		Message: message,
		Frames:  failFrames,
	}
	panic(failure)
}
