package main

import (
	"github.com/armckinney/gyrus/internal/cli"
	_ "github.com/armckinney/gyrus/internal/cli/commands"
)

func main() {
	cli.Execute()
}
