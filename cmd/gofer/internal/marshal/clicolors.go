package marshal

import "strings"

// colorCode represents ANSII escape code for color formatting.
type colorCode string

const (
	reset   colorCode = "\033[0m"
	black   colorCode = "\033[30m"
	red     colorCode = "\033[31m"
	green   colorCode = "\033[32m"
	yellow  colorCode = "\033[33m"
	blue    colorCode = "\033[34m"
	magenta colorCode = "\033[35m"
	cyan    colorCode = "\033[36m"
	white   colorCode = "\033[37m"
)

// color adds given ANSII escape code at beginning of every line.
func color(str string, color colorCode) string {
	return string(color) + strings.ReplaceAll(str, "\n", "\n"+string(reset+color)) + string(reset)
}
