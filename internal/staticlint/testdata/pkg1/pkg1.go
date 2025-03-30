package main

import "os"

func f1() {
	os.Exit(0)
}

func f2() {
	os.Exit(1)
}

func main() {
	os.Exit(0) // want "main function contains os.Exit call"
}
