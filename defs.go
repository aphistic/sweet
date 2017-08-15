package sweet

import (
	"fmt"
	"reflect"
	"testing"
)

func formatName(suiteName string, funcName string) string {
	return fmt.Sprintf("%s/%s", suiteName, funcName)
}

type funcDef struct {
	Name      string
	ParamSets []*paramSet
}

func newFuncDef(name string, paramSets ...*paramSet) *funcDef {
	fd := &funcDef{
		Name:      name,
		ParamSets: make([]*paramSet, 0),
	}

	for _, set := range paramSets {
		fd.ParamSets = append(fd.ParamSets, set)
	}

	return fd
}

func (fd *funcDef) Validate(funcVal reflect.Value) (int, error) {
	if !funcVal.IsValid() || funcVal.Kind() != reflect.Func {
		return 0, errInvalidValue
	}

	typ := funcVal.Type()
	for _, set := range fd.ParamSets {
		// Make sure the number of params match
		if typ.NumIn() != len(set.Params) {
			continue
		}

		// Make sure the params themselves match
		fullMatch := true
		for idx := 0; idx < typ.NumIn(); idx++ {
			if set.Params[idx].Type != typ.In(idx) {
				fullMatch = false
				break
			}
		}

		if fullMatch {
			if set.Deprecated {
				return set.Version, errDeprecated
			}

			return set.Version, nil
		}
	}
	return 0, errUnsupportedMethod
}

type paramSet struct {
	Params     []*paramDef
	Deprecated bool
	Version    int
}

func newParamSet(version int, deprecated bool, params ...*paramDef) *paramSet {
	ps := &paramSet{
		Params:     make([]*paramDef, 0),
		Deprecated: deprecated,
		Version:    version,
	}

	for _, param := range params {
		ps.Params = append(ps.Params, param)
	}

	return ps
}

type paramDef struct {
	Type reflect.Type
}

func newParamDef(typ reflect.Type) *paramDef {
	return &paramDef{Type: typ}
}

var (
	defSetUpAllTests = newFuncDef(
		"SetUpAllTests",
		newParamSet(1, false), // No params
	)
	defTearDownAllTests = newFuncDef(
		"TearDownAllTests",
		newParamSet(1, false), // No params
	)

	defSetUpSuite = newFuncDef(
		"SetUpSuite",
		newParamSet(1, false), // No params
	)
	defTearDownSuite = newFuncDef(
		"TearDownSuite",
		newParamSet(1, false), // No params
	)

	defSetUpTest = newFuncDef(
		"SetUpTest",
		newParamSet(1, true,
			newParamDef(reflect.TypeOf(&testing.T{})),
		),
		newParamSet(2, false,
			newParamDef(reflect.TypeOf((*T)(nil)).Elem()),
		),
	)
	defTearDownTest = newFuncDef(
		"TearDownTest",
		newParamSet(1, true,
			newParamDef(reflect.TypeOf(&testing.T{})),
		),
		newParamSet(2, false,
			newParamDef(reflect.TypeOf((*T)(nil)).Elem()),
		),
	)

	defTest = newFuncDef(
		"Test",
		newParamSet(1, true,
			newParamDef(reflect.TypeOf(&testing.T{})),
		),
		newParamSet(2, false,
			newParamDef(reflect.TypeOf((*T)(nil)).Elem()),
		),
	)
)
