package internal

import (
	"fmt"

	"github.com/fatih/color"
)

var verbose = false

func SetLogVerbosity(value bool) {
	verbose = value
}

var debugC = color.New(color.FgBlue).SprintFunc()("debug")
var infoC = color.New(color.FgGreen).SprintFunc()("info ")
var warningC = color.New(color.FgYellow).SprintFunc()("warn ")
var errorC = color.New(color.FgRed).SprintFunc()("error")

func logDebug(msg string, args ...interface{}) {
	if verbose {
		fmt.Printf("[%s] %s\n", debugC, fmt.Sprintf(msg, args...))
	}
}

func LogInfo(msg string, args ...interface{}) {
	fmt.Printf("[%s] %s\n", infoC, fmt.Sprintf(msg, args...))
}

func LogWarning(msg string, args ...interface{}) {
	fmt.Printf("[%s] %s\n", warningC, fmt.Sprintf(msg, args...))
}

func LogError(msg string, args ...interface{}) {
	fmt.Printf("[%s] %s\n", errorC, fmt.Sprintf(msg, args...))
}
