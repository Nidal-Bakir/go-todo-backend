package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func NewLogger(shouldOutputToConcole bool) zerolog.Logger {
	var output io.Writer

	if shouldOutputToConcole {
		output = zerolog.ConsoleWriter{Out: os.Stdout}
	} else { // its on server. so log to file
		file, err := os.OpenFile(
			"/var/log/todo_proj/todo_proj.log",
			os.O_APPEND|os.O_CREATE|os.O_WRONLY,
			0664,
		)
		if err != nil {
			os.Stderr.Write([]byte(fmt.Sprintf("Error opening the log file for write, Error: %v", err)))
			os.Exit(1)
		}
		output = file
	}

	zerolog.DurationFieldUnit = time.Millisecond
	zerolog.DefaultContextLogger = nil
	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		return filepath.Base(file) + ":" + strconv.Itoa(line)
	}

	log.Logger = zerolog.New(output).With().Caller().Timestamp().Logger()

	return log.Logger
}
