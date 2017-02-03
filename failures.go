package sweet

import (
	"path/filepath"
	"runtime"
)

type testFailure struct {
	Message    string
	Filename   string
	LineNumber int
}

func GomegaFail(message string, callerSkip ...int) {
	filename := "<uknown file, submit a bug report>"
	lineNo := 0
	if len(callerSkip) > 0 {
		_, file, line, ok := runtime.Caller(callerSkip[0] + 1)
		if ok {
			_, name := filepath.Split(file)
			filename = name
			lineNo = line
		}
	}

	failure := &testFailure{
		Message:    message,
		Filename:   filename,
		LineNumber: lineNo,
	}
	panic(failure)
}
