package formatter

import (
	"strings"

	"github.com/sirupsen/logrus"
)

// TextFormatter is a wrapper for the original TextFormatter from the logrus
// package. It removes all fields with the "x-" prefix. This will allow to add
// more data fields to logs without making the CLI output to messy.
type TextFormatter struct {
	f logrus.TextFormatter
}

func (f *TextFormatter) Format(e *logrus.Entry) ([]byte, error) {
	data := logrus.Fields{}
	for k, v := range e.Data {
		if !strings.HasPrefix(k, "x-") {
			data[k] = v
		}
	}
	e.Data = data
	return f.f.Format(e)
}
