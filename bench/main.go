package main

/*
$ sudo launchctl limit maxfiles 1000000 1000000
$ ulimit -n 100000
*/

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/dustin/go-humanize"
)

var (
	URL = "http://127.0.0.1:54321/sub?ch="
)

func doOnce(ok, nok, oops chan bool, etag *string) {
	resp, err := http.Get(URL + *etag)
	if err != nil {
		fmt.Println("Cant load URL", err)
		oops <- true
		return
	}
	if resp.StatusCode != 200 {
		fmt.Println("got non zero code", resp)
		nok <- true
		return
	}
	ok <- true
}

func main() {
	N := flag.Int("n", 100, "Number of parallel connections")
	etag := flag.String("etag", "0", "ETAG")
	flag.Parse()

	start := time.Now()

	ok := make(chan bool)
	nok := make(chan bool)
	oops := make(chan bool)

	for i := 0; i < *N; i++ {
		fmt.Printf("N1 = %d\r", i)
		<-time.After(time.Microsecond * 200)
		go doOnce(ok, nok, oops, etag)
	}
	fmt.Println("")

	fmt.Println("go routines started", *N, URL+*etag, humanize.Time(start))

	n_ok := 0
	n_nok := 0
	n_oops := 0

	for i := 0; i < *N; i++ {
		fmt.Printf("N2 = %d\r", i)
		select {
		case <-ok:
			n_ok += 1
		case <-nok:
			n_nok += 1
		case <-oops:
			n_oops += 1
		}
	}
	fmt.Println("")

	fmt.Println("n_ok", n_ok)
	fmt.Println("n_nok", n_nok)
	fmt.Println("n_oops", n_oops)

	fmt.Println("time: ", humanize.Time(start))

	fmt.Println("\nserver status:")
	resp, err := http.Get("http://localhost:54321/stats")
	if err != nil {
		fmt.Println("cant get status")
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	fmt.Printf("%s", body)
}
