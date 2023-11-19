package main

import (
	"database/sql"
	"log"
	"log/slog"
	"net/http"
	"os"
	"text/template"

	_ "github.com/go-sql-driver/mysql"
	"github.com/rnemeth90/snippetbox/internal/models"
	"github.com/spf13/pflag"
)

var (
	addr string
	dsn  string
)

type application struct {
	logger        *slog.Logger
	snippets      *models.SnippetModel
	templateCache map[string]*template.Template
}

func init() {
	pflag.StringVar(&addr, "addr", ":4000", "address:port to listen on")
	pflag.StringVar(&dsn, "dsn", "web:pass@/snippetbox?parseTime=true", "MySQL DSN")
	pflag.Usage = usage
}

func usage() {
	pflag.PrintDefaults()
}

func main() {
	pflag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	db, err := openDB(dsn)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	defer db.Close()

	templateCache, err := newTemplateCache()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	app := &application{
		logger:        logger,
		templateCache: templateCache,
		snippets:      &models.SnippetModel{DB: db},
	}

	logger.Info("starting server on %s\n", addr)

	err = http.ListenAndServe(addr, app.routes())
	log.Fatal(err)
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
