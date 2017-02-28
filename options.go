package sweet

import "flag"

type optFlags []string

func (f *optFlags) String() string {
	return "string rep"
}

func (f *optFlags) Set(value string) error {
	*f = append(*f, value)
	return nil
}

var (
	flagSkipRuns bool
	flagOpts     optFlags
	flagHelp     = flag.Bool("sweet.help", false, "Shows help information for sweet and registered plugins")
	flagInclude  = flag.String("sweet.include", "", "Only run tests that match the provided expression")
	flagExclude  = flag.String("sweet.exclude", "", "Do not include tests that match the provided expression")
)

func init() {
	flag.Var(&flagOpts, "sweet.opt", "Things")
}
