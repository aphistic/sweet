package sweet

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"testing"

	. "github.com/onsi/gomega"
	"syscall"
)

func TestMain(m *testing.M) {
	RegisterFailHandler(GomegaFail)

	Run(m, func(s *S) {
		s.AddSuite(&RunnerSuite{})
		s.AddSuite(&FailureSuite{})
		s.AddSuite(&ReturnCodeSuite{})
		s.AddSuite(&DefsSuite{})
		s.AddSuite(&TSuite{})

		v1DefSuite := &SweetDefsV1Suite{}
		s.AddSuite(v1DefSuite)
		s.suppressDeprecation(v1DefSuite)
		s.AddSuite(&SweetDefsV2Suite{})

		s.AddSuite(&SuiteV1Suite{})
		s.AddSuite(&SuiteV2Suite{})
	})
}

func runSubTests(name ...string) (int, string, string, error) {
	names := []string{"subtests"}
	names = append(names, name...)
	fullPath := path.Join(names...)

	fi, err := os.Stat(fullPath)
	if err != nil {
		return 0, "", "", err
	}
	if !fi.IsDir() {
		return 0, "", "", fmt.Errorf("%s is not a directory", fullPath)
	}

	cmd := exec.Command("go", "test")
	cmd.Dir = fullPath

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return 0, "", "", err
	}
	defer stdoutPipe.Close()

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return 0, "", "", err
	}
	defer stderrPipe.Close()

	err = cmd.Start()
	if err != nil {
		return 0, "", "", err
	}

	stdout, err := ioutil.ReadAll(stdoutPipe)
	if err != nil {
		return 0, "", "", err
	}

	stderr, err := ioutil.ReadAll(stderrPipe)
	if err != nil {
		return 0, "", "", err
	}

	exitCode := 0
	err = cmd.Wait()
	if err != nil {
		if execErr, ok := err.(*exec.ExitError); ok {
			if waitStatus, ok := execErr.Sys().(syscall.WaitStatus); ok {
				exitCode = waitStatus.ExitStatus()
			}
		}
	}

	return exitCode, string(stdout), string(stderr), nil
}
