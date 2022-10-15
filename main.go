package main

import (
	"os"

	"github.com/go-olive/olive/command"
)

func main() {
	command.Execute(os.Args[1:])
}
