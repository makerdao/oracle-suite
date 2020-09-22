package marshal

import "strings"

type clicolor string

const (
	reset   clicolor = "\033[0m"
	black   clicolor = "\033[30m"
	red     clicolor = "\033[31m"
	green   clicolor = "\033[32m"
	yellow  clicolor = "\033[33m"
	blue    clicolor = "\033[34m"
	magenta clicolor = "\033[35m"
	cyan    clicolor = "\033[36m"
	white   clicolor = "\033[37m"
)

func color(str string, color clicolor) string {
	return string(color) + strings.ReplaceAll(str, "\n", "\n"+string(reset+color)) + string(reset)
}
