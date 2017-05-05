package sweet

import "flag"
import "strings"

type stringSliceFlags []string

func (f stringSliceFlags) String() string {
	return strings.Join(f, ", ")
}

func (f *stringSliceFlags) Set(value string) error {
	*f = append(*f, value)
	return nil
}

var (
	flagSkipRuns bool
	flagOpts     stringSliceFlags
	flagHelp     = flag.Bool("sweet.help", false, "Shows help information for sweet and registered plugins")
	flagExtended = flag.Bool("sweet.extended", false, "Shows extended error information for failed tests")
	flagInclude  stringSliceFlags
	flagExclude  stringSliceFlags
)

func init() {
	flag.Var(&flagOpts, "sweet.opt", "Option to provide to a sweet plugin in the format \"plugin.setting=value\"")
	flag.Var(&flagInclude, "sweet.include", "Only run tests that match the provided expression")
	flag.Var(&flagExclude, "sweet.exclude", "Do not include tests that match the provided expression")
}
