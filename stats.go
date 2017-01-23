package sweet

type stats struct {
	Passed int
	Failed int
}

func newStats() *stats {
	return &stats{}
}
