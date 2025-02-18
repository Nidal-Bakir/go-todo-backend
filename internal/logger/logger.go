package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	colorBlack = iota + 30
	colorRed
	colorGreen
	colorYellow
	colorBlue
	colorMagenta
	colorCyan
	colorWhite
)

func NewLogger(shouldOutputToConcole bool) *zerolog.Logger {
	var output io.Writer

	if shouldOutputToConcole {
		output = zerolog.ConsoleWriter{
			Out:              os.Stdout,
			FormatPrepare:    formatPrepare,
			FormatFieldValue: formatFieldValue,
			FormatFieldName:  formatFieldName,
		}
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

	return &log.Logger
}

// for concole formating
func formatPrepare(m map[string]interface{}) error {

	if v, ok := m["code"]; ok {
		if strNum, ok := v.(json.Number); ok {
			num, err := strconv.Atoi(strNum.String())
			if err == nil {
				color := 0

				switch {
				case num == 200 || num == 201:
					color = colorGreen
				case num >= 400 && num < 500:
					color = colorYellow
				case num >= 500 && num < 600:
					color = colorRed

				}

				if color != 0 {
					m[fmt.Sprintf("-###%d#%v", color, "code")] = fmt.Sprintf("-###%d#%v", color, num)
				}
			}
		}
	}

	return nil
}

// for concole formating
func formatFieldValue(i interface{}) string {
	str := fmt.Sprintf("%s", i)

	if str, ok := strings.CutPrefix(str, "-###"); ok {
		strList := strings.Split(str, "#")
		color := strList[0]
		value := strList[1]
		return fmt.Sprintf("\x1b[%sm%v\x1b[0m", color, value)
	}

	return str
}

// for concole formating
func formatFieldName(i interface{}) string {
	str := fmt.Sprintf("%s=", i)

	if str, ok := strings.CutPrefix(str, "-###"); ok {
		strList := strings.Split(str, "#")
		color := strList[0]
		value := strList[1]
		return fmt.Sprintf("\x1b[%sm%v\x1b[0m", color, value)
	}

	return fmt.Sprintf("\x1b[%dm%v\x1b[0m", colorCyan, str)
}
