package sweet

import (
	"bytes"
	"os"

	"github.com/aphistic/boom"
)

type pipeCapture struct {
	r    *os.File
	w    *os.File
	buf  bytes.Buffer
	task *boom.Task
}

func newPipeCapture() (*pipeCapture, error) {
	r, w, err := os.Pipe()
	if err != nil {
		return nil, err
	}

	pc := &pipeCapture{
		r: r,
		w: w,
	}
	pc.task = boom.RunTask(pc.readPipe)

	pc.task.WaitForRunning(0)

	return pc, nil
}

func (pc *pipeCapture) readPipe(task *boom.Task, args ...interface{}) boom.TaskResult {
	task.SetRunning(true)
	buf := make([]byte, 2048)

readLoop:
	for {
		select {
		case <-task.Stopping():
			break readLoop
		default:
		}

		rN, err := pc.r.Read(buf)
		if err != nil {
			return boom.NewErrorResult(err)
		}

		writeStart := 0

		for {
			wN, err := pc.buf.Write(buf[writeStart : rN-writeStart])
			if err != nil {
				return boom.NewErrorResult(err)
			}
			writeStart += wN
			if writeStart == rN {
				break
			}
		}
	}

	return nil
}

func (pc *pipeCapture) W() *os.File {
	return pc.w
}

func (pc *pipeCapture) Buffer() []byte {
	return pc.buf.Bytes()
}

func (pc *pipeCapture) Close() error {
	err := pc.w.Close()
	if err != nil {
		return err
	}
	err = pc.r.Close()
	if err != nil {
		return err
	}

	return nil
}
