package main

import (
	"github.com/groundsec/waybackshots/cmd"
	"github.com/groundsec/waybackshots/pkg/utils"
)

var version = "0.1.1"

func main() {
	utils.Banner(version)
	cmd.Execute()
}
