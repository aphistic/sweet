package failtests

import (
	"testing"

	"github.com/aphistic/sweet"
	. "github.com/onsi/gomega"
)

func Test(t *testing.T) {
	RegisterFailHandler(sweet.GomegaFail)

	sweet.T(func(s *sweet.S) {
		s.RunSuite(t, &FailSuite{})
	})
}

type FailSuite struct{}

func (s *FailSuite) TestFails(t *testing.T) {
	Expect(false).To(Equal(true))
	Expect("foo").To(Equal("bar"))
}

func (s *FailSuite) TestUtilFunc(t *testing.T) {
	checkTrue := func(val bool) {
		Expect(val).To(BeTrue())
	}

	checkTrue(true)
	checkTrue(true)
	checkTrue(true)
	checkTrue(false)
}
