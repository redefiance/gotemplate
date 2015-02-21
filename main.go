package main

import (
	"flag"
	"go/token"
	"log"
	"os"
	"path"
	"runtime"
)

var fDir = flag.String("d", "", "desc")

var GOPATH = os.Getenv("GOPATH")
var FS = token.NewFileSet()

func main() {
	flag.Parse()

	wd, err := os.Getwd()
	deny(err)

	*fDir = path.Clean(path.Join(wd, *fDir))

	p := parsePackage(*fDir)
	p.update()
}

func deny(err error) {
	if err != nil {
		// log.Fatalln(err)
		panic(err)
	}
}

func assert(condition bool) {
	if condition == false {
		_, file, line, _ := runtime.Caller(1)
		log.Fatalln("assertion failed in %s:%d\n", file, line)
	}
}
