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

type application struct {
	logger *slog.Logger
}

func init() {
	pflag.StringVar(&addr, "addr", ":4000", "address:port to listen on")
	pflag.Usage = usage
}

func usage() {
	pflag.PrintDefaults()
}

func main() {
	pflag.Parse()

	app := application{logger: slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{AddSource: true, Level: slog.LevelDebug}))}

	app.logger.Info("starting server on %s\n", addr)

	err := http.ListenAndServe(addr, app.routes())
	log.Fatal(err)
}
