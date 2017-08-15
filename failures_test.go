package sweet

import (
	. "github.com/onsi/gomega"
)

type FailureSuite struct{}

func (s *FailureSuite) TestIsGoPackage(t T) {
	Expect(isGoPackage("/usr/lib/go/src/runtime/asm_amd64.s")).To(BeTrue())
	Expect(isGoPackage("/usr/lib/go/src/reflect/value.go")).To(BeTrue())

	Expect(isGoPackage("//home/aphistic/go/src/github.com/aphistic/sweet/failtests/failtests_test.go")).To(Not(BeTrue()))
}
