# sweet
Sweet is a pluggable test runner capable of hooking into standard Go tests. It attempts to provide access to the standard Go test tool as close as possible while adding support for test suites and plugins that can hook into test results to add additional functionality.

## Using Sweet

To use Sweet as your test runner you need to add the [TestMain](https://golang.org/pkg/testing/#hdr-Main) function from the Go `testing` package.  This will allow Sweet to run code before and after the actual tests run.  Inside the `TestMain` function you'll want to call the `sweet.Run` function to both set up the Sweet configuration as well as run the tests:

``` Go
package mypackage

import (
    "testing"
    "github.com/aphistic/sweet"
)


func TestMain(m *testing.M) {
    sweet.Run(m, func(s *sweet.S) {
        // Configuration goes here
    })
}
```

## Defining a Suite

A test suite in Sweet is just a normal Go struct with methods named similar to standard go test names (beginning with `Test`).  Creating a suite named `FailSuite` with a single test that always fails and then running it with Sweet looks like the following code:

``` Go
package mypackage

import (
    "testing"
    "github.com/aphistic/sweet"
)


func TestMain(m *testing.M) {
    sweet.Run(m, func(s *sweet.S) {
        s.AddSuite(&FailSuite{})
        // Add any additional suites the same way
    })
}

type FailSuite struct {}

func (s *FailSuite) TestAlwaysFails(t sweet.T) {
    t.Fail()
}
```

## Using a Plugin

Sweet supports plugins to add functionality that isn't typically available with the standard Go testing tools.  One such example is [sweet-junit](https://github.com/aphistic/sweet-junit), a plugin that generates a `junit.xml` file for each package it's used in. To add to the previous examples, this is how you'd add the `sweet-junit` plugin to your tests:

``` Go
package mypackage

import (
    "testing"
    "github.com/aphistic/sweet"
    junit "github.com/aphistic/sweet-junit"
)


func TestMain(m *testing.M) {
    sweet.Run(m, func(s *sweet.S) {
        s.RegisterPlugin(junit.NewPlugin())

        s.AddSuite(&FailSuite{})
        // Add any additional suites the same way
    })
}

type FailSuite struct {}

func (s *FailSuite) TestAlwaysFails(t sweet.T) {
    t.Fail()
}
```

## Using an External Matcher

Sweet was designed with the capability to use external matchers in mind.  You can write standard Go unit tests but you can also hook a different matcher library in and use that.

So far the only matcher that Sweet has been tested with and has hooks for is the [Gomega](https://onsi.github.io/gomega/) library.

To use `Gomega` in the above example, you would do:

``` Go
package mypackage

import (
    "testing"

    . "github.com/onsi/gomega"

    "github.com/aphistic/sweet"
    junit "github.com/aphistic/sweet-junit"
)


func TestMain(m *testing.M) {
    RegisterFailHandler(sweet.GomegaFail)

    sweet.Run(m, func(s *sweet.S) {
        s.RegisterPlugin(junit.NewPlugin())

        s.AddSuite(&FailSuite{})
        // Add any additional suites the same way
    })
}

type FailSuite struct {}

func (s *FailSuite) TestAlwaysFails(t sweet.T) {
    t.Fail()
}
```
