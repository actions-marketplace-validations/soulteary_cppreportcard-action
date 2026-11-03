package main

import (
	"github.com/soulteary/cppreportcard-action/internal/app"
	"os"
)

var version = "dev"

func main() { os.Exit(app.Run(os.Args[1:], os.Stdout, os.Stderr, version)) }
