package main

import (
	"flag"
	"fmt"
	"runtime"
	"time"
)

//go:generate esc -o static.go index.html client.js
//esc: http://godoc.org/github.com/mjibson/esc

func DebugRoutine() {
	for {
		<-time.After(2 * time.Second)
		fmt.Println(time.Now(), "NumGoroutine", runtime.NumGoroutine())
	}
}

func main() {
	flag.Parse()
	if Debug {
		go DebugRoutine()
	}
	ServeHTTP()
}
