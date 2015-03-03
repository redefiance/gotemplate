package main

import (
	"flag"
	"go/token"
	"log"
	"os"
	"path"
	"runtime"
)

var (
	// currently no way to display description of these flags
	fDir       = flag.String("d", "", "TODO")
	fRecursive = flag.Bool("r", false, "TODO")

	GOPATH = os.Getenv("GOPATH")
	FS     = token.NewFileSet()
)

func main() {
	flag.Parse()

	wd, err := os.Getwd()
	deny(err)

	if !path.IsAbs(*fDir) {
		*fDir = path.Join(wd, *fDir)
	}
	path.Clean(*fDir)

	if p := parsePackage(*fDir); p != nil {
		p.generate()
	}
}

func deny(err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		log.Fatalf("Fatal error in %s:%d: %s\n", file, line, err)
	}
}

func assert(condition bool) {
	if condition == false {
		_, file, line, _ := runtime.Caller(1)
		log.Fatalf("assertion failed in %s:%d\n", file, line)
	}
}
