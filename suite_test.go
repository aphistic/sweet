package sweet

import (
	. "github.com/onsi/gomega"
)

type SuiteV1Suite struct{}

func (s *SuiteV1Suite) TestSetUpTearDown(t T) {
	code, stdout, _, err := runSubTests("suitev1", "setupteardown")
	Expect(code).To(Equal(0))
	Expect(err).To(BeNil())

	Expect(stdout).To(ContainSubstring("{SetUpAllTests}\n"))
	Expect(stdout).To(ContainSubstring("{SetUpSuite}\n"))
	Expect(stdout).To(ContainSubstring("{SetUpTest}\n"))
	Expect(stdout).To(ContainSubstring("{TearDownTest}\n"))
	Expect(stdout).To(ContainSubstring("{TearDownAllTests}\n"))
	Expect(stdout).To(ContainSubstring("{TearDownSuite}\n"))
}

type SuiteV2Suite struct{}

func (s *SuiteV2Suite) TestSetUpTearDown(t T) {
	code, stdout, _, err := runSubTests("suitev2", "setupteardown")
	Expect(code).To(Equal(0))
	Expect(err).To(BeNil())

	Expect(stdout).To(ContainSubstring("{SetUpAllTests}\n"))
	Expect(stdout).To(ContainSubstring("{SetUpSuite}\n"))
	Expect(stdout).To(ContainSubstring("{SetUpTest}\n"))
	Expect(stdout).To(ContainSubstring("{TearDownTest}\n"))
	Expect(stdout).To(ContainSubstring("{TearDownAllTests}\n"))
	Expect(stdout).To(ContainSubstring("{TearDownSuite}\n"))
}
