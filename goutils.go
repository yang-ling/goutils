package goutils

import (
	"bytes"
	"io"
	"log"
	"os/exec"
)

type (
	PreErrorHandler  func(err error) (needPanic bool)
	PostErrorHandler func(err error, logger *log.Logger)
)

func ExecCmd(cmd *exec.Cmd, logger *log.Logger) {
	r, w := io.Pipe()
	cmd.Stdout = w
	cmd.Stderr = w
	go func() {
		defer w.Close()
		logger.Println("cmd start")
		HandleError(cmd.Start(), logger)
		HandleError(cmd.Wait(), logger)
		logger.Println("cmd wait")
	}()
	var remaining []byte
	for {
		data := make([]byte, 100)
		n, err := r.Read(data)
		var lines []string
		lines, remaining = writeBytesToLines(data[:n], remaining)
		for _, line := range lines {
			logger.Print(line)
		}
		if err == io.EOF {
			logger.Print(string(remaining))
			break
		} else {
			HandleError(err, logger)
		}
	}
}

func HandleNormalError(err error, logger *log.Logger) {
	preHandler := func(err error) (needPanic bool) {
		return err != nil
	}
	postHandler := func(err error, logger *log.Logger) {}
	HandleError(err, logger, preHandler, postHandler)
}

func HandleError(err error, logger *log.Logger, preHandler PreErrorHandler, postHandler PostErrorHandler) {
	if preHandler(err) {
		postHandler(err, logger)
		logger.Panic(err)
	}
}

func WriteBytesToLines(data, previousRemaining []byte) (lines []string, remaining []byte, err error) {
	var (
		out  bytes.Buffer
		line []byte
	)
	out.Write(data)
	line, err = out.ReadBytes('\n')
	line = append(previousRemaining, line...)
	for err == nil {
		lines = append(lines, string(line))
		line, err = out.ReadBytes('\n')
	}
	if err == io.EOF {
		remaining = make([]byte, len(line))
		copy(remaining, line)
	}
	return
}
