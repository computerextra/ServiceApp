// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package db

import (
	"database/sql"
	"time"
)

type Account struct {
	ID                    string
	Userid                string
	Type                  string
	Provider              string
	Provideraccountid     string
	RefreshToken          sql.NullString
	AccessToken           sql.NullString
	ExpiresAt             sql.NullInt32
	TokenType             sql.NullString
	Scope                 sql.NullString
	IDToken               sql.NullString
	SessionState          sql.NullString
	RefreshTokenExpiresIn sql.NullInt32
}

type Anschprechpartner struct {
	ID            string
	Name          string
	Telefon       sql.NullString
	Mobil         sql.NullString
	Mail          sql.NullString
	Lieferantenid sql.NullString
}

type Aussteller struct {
	Artikelnummer string
	Artikelname   string
	Specs         string
	Preis         string
	Bild          sql.NullString
	ID            int32
}

type Einkauf struct {
	ID            string
	Paypal        bool
	Abonniert     bool
	Geld          sql.NullString
	Pfand         sql.NullString
	Dinge         sql.NullString
	Mitarbeiterid string
	Abgeschickt   sql.NullTime
	Bild1         sql.NullString
	Bild2         sql.NullString
	Bild3         sql.NullString
	Bild1date     sql.NullTime
	Bild2date     sql.NullTime
	Bild3date     sql.NullTime
}

type Fischer struct {
	Username string
	Password string
	Count    int32
}

type Lieferanten struct {
	ID           string
	Firma        string
	Kundennummer sql.NullString
	Webseite     sql.NullString
}

type Mitarbeiter struct {
	ID                 string
	Name               string
	Short              sql.NullString
	Gruppenwahl        sql.NullString
	Interntelefon1     sql.NullString
	Interntelefon2     sql.NullString
	Festnetzalternativ sql.NullString
	Festnetzprivat     sql.NullString
	Homeoffice         sql.NullString
	Mobilbusiness      sql.NullString
	Mobilprivat        sql.NullString
	Email              sql.NullString
	Azubi              sql.NullBool
	Geburtstag         sql.NullTime
}

type Pdf struct {
	Title string
	Body  string
	ID    int32
}

type Session struct {
	ID           string
	Sessiontoken string
	Userid       string
	Expires      time.Time
}

type Short struct {
	Origin string
	Short  string
	Count  sql.NullInt32
	User   sql.NullString
	ID     int32
}

type User struct {
	ID            string
	Name          sql.NullString
	Email         sql.NullString
	Emailverified sql.NullTime
	Image         sql.NullString
	Isadmin       bool
}

type Verificationtoken struct {
	Identifier string
	Token      string
	Expires    time.Time
}

type Warenlieferung struct {
	ID            int32
	Name          string
	Angelegt      time.Time
	Geliefert     sql.NullTime
	Alterpreis    sql.NullString
	Neuerpreis    sql.NullString
	Preis         sql.NullTime
	Artikelnummer string
}

type Wiki struct {
	ID        string
	Name      string
	Inhalt    string
	CreatedAt time.Time
}
