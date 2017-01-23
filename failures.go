package sweet

import (
	"fmt"
)

func GomegaFail(message string, callerSkip ...int) {
	fmt.Printf("Gomega failed\n")
	panic("gomegafail")
}
