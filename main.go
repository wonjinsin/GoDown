package main

import (
	"cheetah/template"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	template.ShowMain()
}
