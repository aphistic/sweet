package sweet

import (
	"bufio"
	"bytes"
	"io"

	"github.com/sergi/go-diff/diffmatchpatch"
)

type differ struct {
	dm *diffmatchpatch.DiffMatchPatch
}

func newDiffer() *differ {
	return &differ{
		dm: diffmatchpatch.New(),
	}
}

func (d *differ) ProcessMessage(message string) string {
	if res, processed := d.gomegaDiff(message); processed {
		return res
	}

	return message
}

func (d *differ) gomegaDiff(message string) (string, bool) {
	var supportedExpectations = []string{
		"to equal",
	}

	msgReader := bufio.NewReader(bytes.NewReader([]byte(message)))

	lineIdx := 0

	insideFirst := false
	firstValue := ""
	insideSecond := false
	secondValue := ""

mainLoop:
	for {
		line, _, err := msgReader.ReadLine()
		if err == io.EOF {
			break
		} else if err != nil {
			return "", false
		}

		if lineIdx == 0 {
			if string(line) != "Expected" {
				// This isn't a gomega error we know
				return "", false
			} else {
				insideFirst = true
				lineIdx++
				continue
			}
		}

		if insideFirst {
			for _, supported := range supportedExpectations {
				if string(line) == supported {
					insideFirst = false
					insideSecond = true
					lineIdx++
					continue mainLoop
				}
			}
			if len(line) > 0 && line[0] != ' ' {
				return "", false
			}

			firstValue += string(line) + "\n"
		}

		if insideSecond {
			secondValue += string(line) + "\n"
		}

		lineIdx++
	}

	diffs := d.dm.DiffMain(firstValue, secondValue, true)
	prettyDiff := d.dm.DiffPrettyText(diffs)

	message += "\nDiff\n" + prettyDiff

	return message, true
}
