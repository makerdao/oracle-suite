package flag

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

// FormattersMap is a map of supported logrus formatters. It is safe to add
// custom formatters to this map.
var FormattersMap = map[string]func() logrus.Formatter{
	"text": func() logrus.Formatter {
		return &logrus.TextFormatter{}
	},
	"json": func() logrus.Formatter {
		return &logrus.JSONFormatter{}
	},
}

// DefaultFormatter is a name of a default formatter. This formatter *must* be
// registered in the FormattersMap map.
var DefaultFormatter = "text"

// FormatTypeValue implements pflag.Value. It represents a flag that allow
// to choose a different logrus formatter.
type FormatTypeValue struct {
	format string
}

// String implements the pflag.Value interface.
func (f *FormatTypeValue) String() string {
	return f.format
}

// Set implements the pflag.Value interface.
func (f *FormatTypeValue) Set(v string) error {
	v = strings.ToLower(v)
	if _, ok := FormattersMap[v]; !ok {
		return fmt.Errorf("unsupported format")
	}
	f.format = v
	return nil
}

// Type implements the pflag.Value interface.
func (f *FormatTypeValue) Type() string {
	return "text|json"
}

// Formatter returns the logrus.Formatter for selected type.
func (f *FormatTypeValue) Formatter() logrus.Formatter {
	if f.format == "" {
		FormattersMap[DefaultFormatter]()
	}
	return FormattersMap[f.format]()
}
