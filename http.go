package main

import (
	"flag"
	"fmt"
	"github.com/amitu/gutils"
	"log"
	"net/http"
	"os"
	"time"
)

var (
	HostPort string
)

func init() {
	flag.StringVar(&HostPort, "http", ":5432", "HTTP Host:Port")
}

func PubHandler(w http.ResponseWriter, r *http.Request) {
	channel := r.FormValue("channel")
	size_s := r.FormValue("size")
	life_s := r.FormValue("life")
	one2one := r.FormValue("one2one") == "true"
	key := r.FormValue("key")

	if channel == "" {
		http.Error(w, "channel is required", http.StatusBadRequest)
		return
	}

	size := 10
	if size_s != "" {
		_, err := fmt.Sscan(size_s, &size)
		if err != nil {
			http.Error(w, "invalid size: "+err.Error(), http.StatusBadRequest)
		}
	}

	life := time.Second * 10
	if life_s != "" {
		_, err := fmt.Sscan(life_s, &life)
		if err != nil {
			http.Error(w, "invalid life: "+err.Error(), http.StatusBadRequest)
		}
	}

	fmt.Fprintf(w, "ok")
}

func OKHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "ok")
}

func ServeHTTP() {
	http.HandleFunc("/", OKHandler)
	http.HandleFunc("/pub", PubHandler)
	http.HandleFunc("/sub", OKHandler)
	http.HandleFunc("/stats", OKHandler)
	http.HandleFunc("/client.js", OKHandler)
	http.HandleFunc("/iframe.html", OKHandler)

	log.Printf("Started HTTP Server on %s.", HostPort)
	logger := gutils.NewApacheLoggingHandler(http.DefaultServeMux, os.Stderr)
	server := &http.Server{
		Addr:    HostPort,
		Handler: logger,
	}
	log.Fatal(server.ListenAndServe())
}
