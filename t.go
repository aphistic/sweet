package sweet

import (
	"fmt"
	"sync"
	"testing"
)

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
}

var _ T = &sweetT{}

type sweetT struct {
	t *testing.T

	output []string

	lock sync.RWMutex

	skipped bool
	failed  bool
}

func newSweetT(t *testing.T) *sweetT {
	return &sweetT{
		t:      t,
		output: make([]string, 0),
	}
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
	t.output = append(t.output, fmt.Sprint(args...))
}
func (t *sweetT) Logf(format string, args ...interface{}) {
	t.output = append(t.output, fmt.Sprintf(format, args...))
}

func (t *sweetT) Name() string {
	return t.t.Name()
}

func (t *sweetT) Parallel() {
	t.t.Parallel()
}

func (t *sweetT) Run(name string, f func(t T)) bool {
	panic("Run on sweet.T is not supported yet")
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
