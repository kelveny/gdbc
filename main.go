package main

import (
	"fmt"
	"os"

	"github.com/kelveny/gdbc/cmd"
)

var SemVer = "v0.0.0-devel"

func GetSemverInfo() string {
	return SemVer
}

func main() {
	if len(os.Args) > 1 {
		if os.Args[1] == "version" {
			fmt.Printf("%s\n", GetSemverInfo())
			os.Exit(0)
		}
	}

	cmd.Execute()
}
