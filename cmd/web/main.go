package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/spf13/pflag"
)

var (
	addr string
)

func init() {
	pflag.StringVar(&addr, "addr", ":4000", "address:port to listen on")
	pflag.Usage = usage
}

func usage() {
	pflag.PrintDefaults()
}

func main() {
	pflag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{AddSource: true, Level: slog.LevelDebug}))

	s := myObject{some: "Hello"}
	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("./ui/static/"))

	mux.Handle("/static/", http.StripPrefix("/static", fs))

	mux.HandleFunc("/", home)
	mux.HandleFunc("/snippet/view", snippetView)
	mux.HandleFunc("/snippet/create", snippetCreate)
	mux.Handle("/s", &s)

	log.Printf("starting server on %s\n", addr)
	logger.Info("starting server on %s\n", addr)

	err := http.ListenAndServe(addr, mux)
	log.Fatal(err)
}
