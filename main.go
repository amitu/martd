package main

import (
	"flag"
	"fmt"
	"runtime"
	"time"
)

// esc -o static.go static/

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
