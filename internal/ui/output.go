package ui

import (
	"fmt"

	"github.com/fatih/color"
)

// Success prints a success message in green bold.
func Success(msg string) {
	c := color.New(color.FgGreen, color.Bold)
	c.Printf("%s\n", msg)
}

// Successf prints a formatted success message in green bold.
func Successf(format string, a ...interface{}) {
	c := color.New(color.FgGreen, color.Bold)
	c.Printf(format+"\n", a...)
}

// Error prints an error message in red.
func Error(msg string) {
	c := color.New(color.FgRed)
	c.Printf("Error: %s\n", msg)
}

// Errorf prints a formatted error message in red.
func Errorf(format string, a ...interface{}) {
	c := color.New(color.FgRed)
	c.Printf(format+"\n", a...)
}

// Info prints an informational message in cyan.
func Info(msg string) {
	c := color.New(color.FgCyan)
	c.Printf("%s\n", msg)
}

// Usage prints usage/help text in yellow.
func Usage(msg string) {
	c := color.New(color.FgYellow)
	c.Printf("%s\n", msg)
}

// Println a simple wrapper that prints without color when needed.
func Println(a ...interface{}) {
	fmt.Println(a...)
}
