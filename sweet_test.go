package sweet

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestMain(m *testing.M) {
	RegisterFailHandler(GomegaFail)

	Run(m, func(s *S) {
		s.AddSuite(&RunnerSuite{})
		s.AddSuite(&FailureSuite{})
		s.AddSuite(&ReturnCodeSuite{})
	})
}
