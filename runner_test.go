package sweet

import (
	"os/exec"
	"syscall"

	. "github.com/onsi/gomega"
)

type ReturnCodeSuite struct{}

func (s *ReturnCodeSuite) TestFailureReturnCode(t T) {
	// This is SUPER meta and weird.  We're actually going to run
	// "go test" on the "failtests" directory and make sure it returns
	// a non-zero status code
	cmd := exec.Command("go", "test")
	cmd.Dir = "failtests"
	err := cmd.Run()

	execErr, ok := err.(*exec.ExitError)
	Expect(ok).To(BeTrue())

	waitStatus, ok := execErr.Sys().(syscall.WaitStatus)
	Expect(ok).To(BeTrue())

	Expect(waitStatus.ExitStatus()).To(Equal(1))
}

type RunnerSuite struct{}

func (s *RunnerSuite) TestFailInGoroutines(t T) {
	/*ch := make(chan struct{})
	go func() {
		defer close(ch)
		t.FailNow()
	}()
	<-ch*/
}

func (s *RunnerSuite) TestGomegaFailInGoroutines(t T) {
	/*ch := make(chan struct{})
	go func() {
		defer close(ch)
		Expect(false).To(BeTrue())
	}()
	<-ch*/
}
