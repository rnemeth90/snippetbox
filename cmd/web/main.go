package main

import (
	"crypto/tls"
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"text/template"
	"time"

	"github.com/alexedwards/scs/mysqlstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
	_ "github.com/go-sql-driver/mysql"
	"github.com/rnemeth90/snippetbox/internal/models"
	"github.com/spf13/pflag"
)

var (
	addr           string
	dsn            string
	debugMode bool
)

type application struct {
	logger         *slog.Logger
	snippets       models.SnippetModelInterface
	users          models.UserModelInterface
	templateCache  map[string]*template.Template
	formDecoder    *form.Decoder
	sessionManager *scs.SessionManager
	debugMode           bool
}

func init() {
	pflag.StringVar(&addr, "addr", ":4000", "address:port to listen on")
	pflag.BoolVar(&debugMode, "debug", false, "debug mode")
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

	formDecoder := form.NewDecoder()

	sessionManager := scs.New()
	sessionManager.Store = mysqlstore.New(db)
	sessionManager.Lifetime = 12 * time.Hour
	sessionManager.Cookie.Secure = true

	app := &application{
		logger:         logger,
		templateCache:  templateCache,
		users:          &models.UserModel{DB: db},
		snippets:       &models.SnippetModel{DB: db},
		formDecoder:    formDecoder,
		sessionManager: sessionManager,
		debugMode: debugMode,
	}

	tlsConfig := &tls.Config{
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
	}

	server := http.Server{
		Addr:         addr,
		Handler:      app.routes(),
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		TLSConfig:    tlsConfig,
	}

	logger.Info("starting server", "addr", server.Addr)

	err = server.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
	logger.Error(err.Error())
	os.Exit(1)
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
