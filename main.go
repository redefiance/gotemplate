package main

import (
	"flag"
	"os"
)

var fDir = flag.String("d", "", "desc")

func main() {
	if *fDir == "" {
		wd, err := os.Getwd()
		deny(err)
		*fDir = wd
	}

	parse()
	print()
}

func deny(err error) {
	if err != nil {
		// log.Fatalln(err)
		panic(err)
	}
}
