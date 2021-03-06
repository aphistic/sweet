package sweet

import (
	"fmt"
	"io/ioutil"
	"sync"
	"testing"
)

type SweetUtil interface {
	ListFiles(path string) []string
	LoadFile(path string) []byte
}

type sweetUtil struct {
	t T
}

func (u *sweetUtil) ListFiles(path string) []string {
	nodes, err := ioutil.ReadDir(path)
	if err != nil {
		failTest(err.Error(), 0)
	}

	res := []string{}
	for _, node := range nodes {
		if node.IsDir() {
			continue
		}

		res = append(res, node.Name())
	}

	return res
}

func (u *sweetUtil) LoadFile(path string) []byte {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		failTest(err.Error(), 0)
	}

	return data
}

type T interface {
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Fail()
	FailNow()
	Failed() bool
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	Log(args ...interface{})
	Logf(format string, args ...interface{})
	Name() string
	Parallel()
	Run(name string, f func(t T)) bool
	Skip(args ...interface{})
	SkipNow()
	Skipf(format string, args ...interface{})
	Skipped() bool

	Sweet() SweetUtil
}

var _ T = &sweetT{}

type sweetT struct {
	t    *testing.T
	name *TestName

	logLock sync.RWMutex
	output  []string

	lock sync.RWMutex

	skipped bool
	failed  bool

	util *sweetUtil
}

func newSweetT(t *testing.T, name *TestName) *sweetT {
	newT := &sweetT{
		t:    t,
		name: name,

		output: make([]string, 0),
	}
	newT.util = &sweetUtil{
		t: newT,
	}

	return newT
}

func (t *sweetT) Error(args ...interface{}) {
	t.Log(args...)
	t.Fail()
}
func (t *sweetT) Errorf(format string, args ...interface{}) {
	t.Logf(format, args...)
	t.Fail()
}

func (t *sweetT) Fail() {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.t.Fail()
	t.failed = true
}
func (t *sweetT) FailNow() {
	t.Fail()
	failTest("", 2)
}
func (t *sweetT) Failed() bool {
	t.lock.RLock()
	defer t.lock.RUnlock()

	return t.failed
}

func (t *sweetT) Fatal(args ...interface{}) {
	t.Fail()
	failTest(fmt.Sprint(args...), 1)
}
func (t *sweetT) Fatalf(format string, args ...interface{}) {
	t.Fail()
	failTest(fmt.Sprintf(format, args...), 1)
}

func (t *sweetT) Log(args ...interface{}) {
	t.logLock.Lock()
	defer t.logLock.Unlock()

	t.output = append(t.output, fmt.Sprint(args...))
}
func (t *sweetT) Logf(format string, args ...interface{}) {
	t.logLock.Lock()
	defer t.logLock.Unlock()

	t.output = append(t.output, fmt.Sprintf(format, args...))
}

func (t *sweetT) Name() string {
	return t.name.String()
}

func (t *sweetT) Parallel() {
	t.t.Parallel()
}

func (t *sweetT) Run(name string, f func(t T)) bool {
	var panicValue interface{}

	subName := t.name.Clone()
	subName.AddTestName(name)

	runRes := t.t.Run(name, func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				panicValue = r
			}
		}()
		f(newSweetT(t, subName))
	})

	if panicValue != nil {
		if pvTestFailed, ok := panicValue.(*testFailed); ok {
			if pvTestFailed.TestName == nil {
				pvTestFailed.TestName = subName
			}
		}
		panic(panicValue)
	}

	return runRes
}

func (t *sweetT) Skip(args ...interface{}) {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.skipped = true
	skipTest(fmt.Sprint(args...))
}
func (t *sweetT) SkipNow() {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.skipped = true
	skipTest("")
}
func (t *sweetT) Skipf(format string, args ...interface{}) {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.skipped = true
	skipTest(fmt.Sprintf(format, args...))
}
func (t *sweetT) Skipped() bool {
	t.lock.RLock()
	defer t.lock.RUnlock()

	return t.skipped
}

func (t *sweetT) Sweet() SweetUtil {
	return t.util
}
