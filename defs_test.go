package sweet

import (
	"testing"

	. "github.com/onsi/gomega"
	"reflect"
)

type DefsSuite struct{}

func (s *DefsSuite) TestValidateDeprecated(t T) {
	testDef := newFuncDef("MyFunc",
		newParamSet(1, true,
			newParamDef(reflect.TypeOf(&testing.T{})),
		),
	)

	v, err := testDef.Validate(reflect.ValueOf(func(t *testing.T) {}))
	Expect(v).To(Equal(1))
	Expect(err).To(Equal(errDeprecated))
}

func (s *DefsSuite) TestValidateValid(t T) {
	testDef := newFuncDef("MyFunc",
		newParamSet(1, true,
			newParamDef(reflect.TypeOf(&testing.T{})),
		),
		newParamSet(2, false,
			newParamDef(reflect.TypeOf((*T)(nil)).Elem()),
		),
	)

	v, err := testDef.Validate(reflect.ValueOf(func(t T) {}))
	Expect(v).To(Equal(2))
	Expect(err).To(BeNil())
}

func (s *DefsSuite) TestValidateUnsupportedMethod(t T) {
	testDef := newFuncDef("MyFunc",
		newParamSet(1, false,
			newParamDef(reflect.TypeOf(&testing.T{})),
		),
	)

	// Make sure a param count mismatch fails
	v, err := testDef.Validate(reflect.ValueOf(func() {}))
	Expect(v).To(Equal(0))
	Expect(err).To(Equal(errUnsupportedMethod))

	// Make sure a param type mismatch fails
	v, err = testDef.Validate(reflect.ValueOf(func(s string) {}))
	Expect(v).To(Equal(0))
	Expect(err).To(Equal(errUnsupportedMethod))
}

// Make sure the v1 parameters are passed in correctly
type SweetDefsV1Suite struct{}

func (s *SweetDefsV1Suite) SetUpTest(t *testing.T) {
	Expect(t).ToNot(BeNil())
}
func (s *SweetDefsV1Suite) TestRealTest(t *testing.T) {
	Expect(t).ToNot(BeNil())
}
func (s *SweetDefsV1Suite) TearDownTest(t *testing.T) {
	Expect(t).ToNot(BeNil())
}

// Make sure the v2 parameters are passed in correctly
type SweetDefsV2Suite struct{}

func (s *SweetDefsV2Suite) SetUpTest(t T) {
	Expect(t).ToNot(BeNil())

	tType := reflect.TypeOf(t)
	expectType := reflect.TypeOf(&sweetT{})
	if tType != expectType {
		panic("SweetDefsV2Suite/SetUpTest param is not correct")
	}
}
func (s *SweetDefsV2Suite) TestRealTest(t T) {
	Expect(t).ToNot(BeNil())

	tType := reflect.TypeOf(t)
	expectType := reflect.TypeOf(&sweetT{})
	Expect(tType).To(Equal(expectType))
}

func (s *SweetDefsV2Suite) TearDownTest(t T) {
	Expect(t).ToNot(BeNil())

	tType := reflect.TypeOf(t)
	expectType := reflect.TypeOf(&sweetT{})
	if tType != expectType {
		panic("SweetDefsV2Suite/TearDownTest param is not correct")
	}
}
