package sweet

import (
	. "github.com/onsi/gomega"
)

type differSuite struct{}

func (s *differSuite) TestGomegaStruct(t T) {
	failMessage := `Expected
    <*failtests.testStruct | 0xc00000e880>: {
        StringValue: "this is a string",
        IntValue: 1234,
        BoolValue: true,
    }
to equal
    <*failtests.testStruct | 0xc00000e8a0>: {
        StringValue: "this is not a string",
        IntValue: 1324,
        BoolValue: false,
    }`

	d := newDiffer()
	res := d.ProcessMessage(failMessage)

	Expect(res).To(Equal(`Expected
    <*failtests.testStruct | 0xc00000e880>: {
        StringValue: "this is a string",
        IntValue: 1234,
        BoolValue: true,
    }
to equal
    <*failtests.testStruct | 0xc00000e8a0>: {
        StringValue: "this is not a string",
        IntValue: 1324,
        BoolValue: false,
    }
Diff
    <*failtests.testStruct | 0xc00000e8[31m8[0m[32ma[0m0>: {
        StringValue: "this is [32mnot [0ma string",
        IntValue: 1[31m2[0m3[32m2[0m4,
        BoolValue: [31mtru[0m[32mfals[0me,
    }
`))
}

func (s *differSuite) TestGomegaString(t T) {
	failMessage := `Expected
    <string>: 
        this
    is
    		a
    string?
to equal
    <string>: 
    this
    	is	
    		a	string
    	`

	d := newDiffer()
	res := d.ProcessMessage(failMessage)

	// 	fmt.Printf("%s\n", res)

	Expect(res).To(Equal(`Expected
    <string>: 
        this
    is
    		a
    string?
to equal
    <string>: 
    this
    	is	
    		a	string
    	
Diff
    <string>: 
    [31m    [0mthis
    [32m	[0mis[32m	[0m
    		a[31m
    [0m[32m	[0mstring[31m?[0m[32m
    	[0m
`))
}
