package main

import "os"

func f1() {
	os.Exit(0)
}

func main() {
	os.Exit(0) // want "main function contains os.Exit call"
}

func f2() {
	os.Exit(0)
}
