package sweet

import (
	"strings"
)

type TestName struct {
	SuiteName string
	TestNames []string
}

func newTestName(suite string, test []string) *TestName {
	return &TestName{
		SuiteName: suite,
		TestNames: test,
	}
}

func (tn *TestName) String() string {
	testName := make([]string, 0, len(tn.TestNames)+1)
	testName = append(testName, tn.SuiteName)
	testName = append(testName, tn.TestNames...)
	return strings.Join(testName, "/")
}

func (tn *TestName) Clone() *TestName {
	newName := &TestName{
		SuiteName:tn.SuiteName,
		TestNames: []string{},
	}
	for _, testName := range tn.TestNames {
		newName.TestNames = append(newName.TestNames, testName)
	}

	return newName
}

func (tn *TestName) AddTestName(testName string) {
	tn.TestNames = append(tn.TestNames, testName)
}
