package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/caarlos0/env/v6"
	_ "github.com/denisenkom/go-mssqldb"
	"github.com/joho/godotenv"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

func (a *App) GetSeriennummer(Artikelnummer string) string {
	database, err := sql.Open("sqlserver", getSageConnectionString())
	if err != nil {
		fehler := err.Error()
		return fehler
	}
	defer database.Close()

	rows, err := database.Query(fmt.Sprintf("SELECT SUCHBEGRIFF FROM sg_auf_artikel WHERE ARTNR LIKE '%s';", Artikelnummer))
	if err != nil {
		fehler := err.Error()
		return fehler
	}
	defer rows.Close()
	var Suchbegriff string
	for rows.Next() {
		if err := rows.Scan(&Suchbegriff); err != nil {
			fehler := err.Error()
			return fehler
		}
	}
	if err := rows.Err(); err != nil {
		fehler := err.Error()
		return fehler
	}
	return Suchbegriff
}

func getSageConnectionString() string {
	env := getEnv()
	server := env.SAGE_SERVER
	db := env.SAGE_DB
	user := env.SAGE_USER
	password := env.SAGE_PASS
	port := env.SAGE_PORT

	return fmt.Sprintf("server=%s;database=%s;user id=%s;password=%s;port=%d", server, db, user, password, port)
}

type Config struct {
	DATABASE_URL     string `env:"DATABASE_URL,required"`
	CMS_DATABASE_URL string `env:"CMS_DATABASE_URL,required"`
	ARCHIVE_PATH     string `env:"ARCHIVE_PATH,required"`
	MAIL_FROM        string `env:"MAIL_FROM,required"`
	MAIL_SERVER      string `env:"MAIL_SERVER,required"`
	MAIL_PORT        int    `env:"MAIL_PORT,required"`
	MAIL_USER        string `env:"MAIL_USER,required"`
	MAIL_PASSWORD    string `env:"MAIL_PASSWORD,required"`
	SAGE_SERVER      string `env:"SAGE_SERVER,required"`
	SAGE_PORT        int    `env:"SAGE_PORT,required"`
	SAGE_USER        string `env:"SAGE_USER,required"`
	SAGE_PASS        string `env:"SAGE_PASS,required"`
	SAGE_DB          string `env:"SAGE_DB,required"`
	ACCESS_DB        string `env:"ACCESS_DB,required"`
}

// Get .env Vars
func getEnv() Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("unable to load .env file: %e", err)
	}

	cfg := Config{} // 👈 new instance of `Config`

	err = env.Parse(&cfg) // 👈 Parse environment variables into `Config`
	if err != nil {
		log.Fatalf("unable to parse ennvironment variables: %e", err)
	}

	return cfg
}
