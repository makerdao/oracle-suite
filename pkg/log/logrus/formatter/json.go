package formatter

import "github.com/sirupsen/logrus"

// JSONFormatter is a wrapper for the original JSONFormatter from the logrus
// package. It changes the timezone of logs to UTC.
type JSONFormatter struct {
	f logrus.JSONFormatter
}

func (f *JSONFormatter) Format(e *logrus.Entry) ([]byte, error) {
	e.Time = e.Time.UTC()
	return f.f.Format(e)
}
