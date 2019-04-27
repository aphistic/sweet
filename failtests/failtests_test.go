package failtests

import (
	"testing"

	"github.com/aphistic/sweet"
	. "github.com/onsi/gomega"
)

func TestMain(m *testing.M) {
	RegisterFailHandler(sweet.GomegaFail)

	sweet.Run(m, func(s *sweet.S) {
		s.AddSuite(&FailSuite{})
	})
}

type FailSuite struct{}

func (s *FailSuite) TestFails(t sweet.T) {
	Expect(false).To(Equal(true))
	Expect("foo").To(Equal("bar"))
}

func (s *FailSuite) TestUtilFunc(t sweet.T) {
	checkTrue := func(val bool) {
		Expect(val).To(BeTrue())
	}

	checkTrue(true)
	checkTrue(true)
	checkTrue(true)
	checkTrue(false)
}

type testStruct struct {
	StringValue string
	IntValue    int
	BoolValue   bool
}

func (s *FailSuite) TestStruct(t sweet.T) {
	Expect(&testStruct{
		StringValue: "this is a string",
		IntValue:    1234,
		BoolValue:   true,
	}).To(Equal(&testStruct{
		StringValue: "this is not a string",
		IntValue:    1324,
		BoolValue:   false,
	}))
}

func (s *FailSuite) TestLongString(t sweet.T) {
	Expect(`
    this
is
		a
string?`).To(Equal(`
this
	is	
		a	string
	`))
}
