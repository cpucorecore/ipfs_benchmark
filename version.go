package main

import "fmt"

const Version = "0.0.3"

var (
	commit  string
	builtAt string
	builtBy string
	builtOn string
)

func version() {
	fmt.Println(Version)
	fmt.Printf("commit: %s ", commit)
	fmt.Printf("built @ %s by %s on %s\n", builtAt, builtBy, builtOn)
}
