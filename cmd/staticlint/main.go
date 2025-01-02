package main

import (
	"metrics/internal/staticlint"

	"golang.org/x/tools/go/analysis/multichecker"
)

func main() {
	multichecker.Main(
		staticlint.Analyzers()...,
	)
}
