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

func GomegaFail(message string, callerSkip ...int) {
	/*
		debug.PrintStack()

		fmt.Printf("Caller skip: %#v\n", callerSkip)

		idx := 0
		for {
			_, file, line, ok := runtime.Caller(idx)
			fmt.Printf("Caller - %s:%d\n", file, line)
			if !ok {
				break
			}
			idx++
		}*/

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
