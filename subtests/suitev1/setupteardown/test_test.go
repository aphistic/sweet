package setupteardown

import (
	"fmt"
	"testing"

	"github.com/aphistic/sweet"
)

func TestMain(m *testing.M) {
	sweet.Run(m, func(s *sweet.S) {
		s.SetUpAllTests(func(t sweet.T) {
			fmt.Printf("{SetUpAllTests}\n")
		})
		s.TearDownAllTests(func(t sweet.T) {
			fmt.Printf("{TearDownAllTests}\n")
		})

		s.AddSuite(&RunSuite{})
	})
}

type RunSuite struct{}

func (s *RunSuite) SetUpSuite() {
	fmt.Printf("{SetUpSuite}\n")
}
func (s *RunSuite) TearDownSuite() {
	fmt.Printf("{TearDownSuite}\n")
}

func (s *RunSuite) SetUpTest(t *testing.T) {
	fmt.Printf("{SetUpTest}\n")
}
func (s *RunSuite) TearDownTest(t *testing.T) {
	fmt.Printf("{TearDownTest}\n")
}

func (s *RunSuite) TestSetUps(t *testing.T) {}
