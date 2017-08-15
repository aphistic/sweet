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
		s.AddSuite(&DefsSuite{})
		s.AddSuite(&TSuite{})

		v1DefSuite := &SweetDefsV1Suite{}
		s.AddSuite(v1DefSuite)
		s.suppressDeprecation(v1DefSuite)
		s.AddSuite(&SweetDefsV2Suite{})
	})
}
