// Package main provides the entry point for the siyuan CLI.
package main

import (
	"os"

	"siyuan/internal/human"
)

func main() {
	os.Exit(human.Execute(os.Args[1:], os.Stdin, os.Stdout))
}
