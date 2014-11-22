package main

import (
	"flag"
)

// esc -o static.go static/

func main() {
	flag.Parse()
	ServeHTTP()
}
