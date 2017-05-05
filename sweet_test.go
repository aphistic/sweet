package sweet

import (
	"testing"

	. "github.com/onsi/gomega"
)

func Test(t *testing.T) {
	RegisterFailHandler(GomegaFail)

	T(func(s *S) {
		s.RunSuite(t, &RunnerSuite{})
		s.RunSuite(t, &FailureSuite{})
		s.RunSuite(t, &ReturnCodeSuite{})
	})
}
