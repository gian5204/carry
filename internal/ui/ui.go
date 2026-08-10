package ui

const (
	bold  = "\x1b[1m"
	dim   = "\x1b[2m"
	green = "\x1b[32m"
	reset = "\x1b[0m"
)

func Bold(text string) string {
	return bold + text + reset
}

func Dim(text string) string {
	return dim + text + reset
}

func Green(text string) string {
	return green + text + reset
}
