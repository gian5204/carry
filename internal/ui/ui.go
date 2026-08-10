package ui

const (
	bold   = "\x1b[1m"
	cyan   = "\x1b[36m"
	dim    = "\x1b[2m"
	green  = "\x1b[32m"
	yellow = "\x1b[33m"
	reset  = "\x1b[0m"
)

func Bold(text string) string {
	return bold + text + reset
}

func Dim(text string) string {
	return dim + text + reset
}

func Cyan(text string) string {
	return cyan + text + reset
}

func Green(text string) string {
	return green + text + reset
}

func BoldGreen(text string) string {
	return bold + green + text + reset
}

func Yellow(text string) string {
	return yellow + text + reset
}
