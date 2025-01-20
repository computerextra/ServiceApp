package main

import (
	"ServiceApp/config"
	"ServiceApp/db"
	"bytes"
	"context"
	"crypto/tls"
	"database/sql"
	"flag"
	"fmt"
	"html/template"
	"log"
	"slices"
	"strconv"
	"strings"
	"time"

	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-adodb"
	gomail "gopkg.in/mail.v2"
)

type Artikel struct {
	Id            int
	Artikelnummer string
	Suchbegriff   string
}

type Leichen struct {
	Artikelnummer string
	Artikelname   string
	Bestand       int16
	Verfügbar     int16
	EK            float64
	LetzterUmsatz string
}

type SummenArtikel struct {
	Artikelnummer string
	Artikelname   string
	Bestand       int16
	EK            float64
	Summe         float64
}

type VerfArtikel struct {
	SummenArtikel
	Verfügbar int16
}

type WertArtikel struct {
	Bestand   int16
	Verfügbar int16
	EK        float64
}

type History struct {
	Id     int
	Action string
}

type Price struct {
	Id     int
	Action string
	Price  float32
}

type AlteSeriennummer struct {
	ArtNr       string
	Suchbegriff string
	Bestand     int
	Verfügbar   int
	GeBeginn    string
}

type AccessArtikel struct {
	Id            int
	Artikelnummer string
	Artikeltext   string
	Preis         float64
}
type AusstellerArtikel struct {
	Id            int
	Artikelnummer string
	Artikelname   string
	Specs         string
	Preis         float64
}
type SageArtikel struct {
	Id            int
	Artikelnummer string
	Suchbegriff   string
	Preis         float64
}

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

func (a *App) SyncAussteller() string {
	// TODO: Machen das es geht, rödelt sich zu tode
	err := syncAussteller(a.ctx)
	if err != nil {
		fehler := err.Error()
		return fehler
	}
	return "OK"
}

func (a *App) SyncLabel() string {
	sageItems, err := readSage()
	if err != nil {
		fehler := err.Error()

		return fehler
	}
	label, err := readAccessDb()
	if err != nil {
		fehler := err.Error()

		return fehler
	}
	err = syncDb(sageItems, label)
	if err != nil {
		fehler := err.Error()

		return fehler
	}
	return "OK"
}

func (a *App) GenerateWarenlieferung() string {

	ctx := a.ctx
	env := config.GetEnv()

	var err error

	datebase, err := sql.Open("mysql", env.DATABASE_URL)
	if err != nil {
		panic(err)
	}
	datebase.SetConnMaxIdleTime(time.Minute * 3)
	datebase.SetMaxOpenConns(10)
	datebase.SetMaxIdleConns(10)
	queries := db.New(datebase)
	Products, err := queries.GetWarenlieferung(ctx)
	if err != nil {
		fehler := err.Error()
		return fehler
	}

	neueArtikel, geliefert, neuePreise, err := sortProducts(Products)
	if err != nil {
		fehler := err.Error()
		return fehler
	}

	for i := range neueArtikel {
		_, err := queries.InsertWarenlieferung(ctx, db.InsertWarenlieferungParams{
			ID:            neueArtikel[i].ID,
			Name:          neueArtikel[i].Name,
			Artikelnummer: neueArtikel[i].Artikelnummer,
		})
		if err != nil {
			fehler := err.Error()
			return fehler
		}
	}
	for i := range geliefert {
		_, err := queries.UpdateWarenlieferung(ctx, db.UpdateWarenlieferungParams{
			Name: geliefert[i].Name,
			ID:   geliefert[i].ID,
		})
		if err != nil {
			fehler := err.Error()
			return fehler
		}
	}
	for i := range neuePreise {
		var altFloat float64
		var neuFloat float64
		if neuePreise[i].Alterpreis.Valid {
			altFloat, _ = strconv.ParseFloat(neuePreise[i].Alterpreis.String, 64)
		}
		if neuePreise[i].Neuerpreis.Valid {
			neuFloat, _ = strconv.ParseFloat(neuePreise[i].Neuerpreis.String, 64)
		}
		if neuFloat > 0 && altFloat > 0 && altFloat != neuFloat {
			_, err := queries.UpdatePreisWarenlieferung(ctx, db.UpdatePreisWarenlieferungParams{
				Alterpreis: neuePreise[i].Alterpreis,
				Neuerpreis: neuePreise[i].Neuerpreis,
				ID:         neuePreise[i].ID,
			})
			if err != nil {
				fehler := err.Error()
				return fehler
			}
		}

	}
	datebase.Close()

	if err != nil {
		fehler := err.Error()
		return fehler
	} else {
		return "OK"
	}
}

func (a *App) SendWarenlieferung() string {

	ctx := a.ctx
	env := config.GetEnv()
	datebase, err := sql.Open("mysql", env.DATABASE_URL)
	if err != nil {
		panic(err)
	}
	datebase.SetConnMaxIdleTime(time.Minute * 3)
	datebase.SetMaxOpenConns(10)
	datebase.SetMaxIdleConns(10)
	queries := db.New(datebase)

	Mitarbeiter, err := queries.GetUsers(ctx)
	if err != nil {
		fehler := err.Error()
		return fehler
	}
	neueArtikel, err := queries.GetDailyNew(ctx)
	if err != nil {
		fehler := err.Error()
		return fehler
	}
	gelieferteArtikel, err := queries.GetDailyDelivered(ctx)
	if err != nil {
		fehler := err.Error()
		return fehler
	}
	neuePreise, err := queries.GetDailyWarenlieferung(ctx)
	if err != nil {
		fehler := err.Error()
		return fehler
	}
	wertBestand, wertVerfügbar, err := getLagerWert()
	if err != nil {
		fehler := err.Error()
		return fehler
	}
	teureArtikel, err := getHighestSum()
	if err != nil {
		fehler := err.Error()
		return fehler
	}
	teureVerfArtikel, err := getHighestVerfSum()
	if err != nil {
		fehler := err.Error()
		return fehler
	}
	leichen, err := getLeichen()
	if err != nil {
		fehler := err.Error()
		return fehler
	}
	SN, err := getAlteSeriennummern()
	if err != nil {
		fehler := err.Error()
		return fehler
	}

	MAIL_FROM := env.MAIL_FROM
	MAIL_SERVER := env.MAIL_SERVER
	MAIL_PORT := env.MAIL_PORT
	MAIL_USER := env.MAIL_USER
	MAIL_PASSWORD := env.MAIL_PASSWORD

	var body string
	if len(neueArtikel) > 0 {
		body = fmt.Sprintf("%s<h2>Neue Artikel</h2><ul>", body)

		for i := range neueArtikel {
			body = fmt.Sprintf("%s<li><b>%s</b> - %s</li>", body, neueArtikel[i].Artikelnummer, neueArtikel[i].Name)
		}
		body = fmt.Sprintf("%s</ul>", body)
	}

	if len(gelieferteArtikel) > 0 {
		body = fmt.Sprintf("%s<br><br><h2>Gelieferte Artikel</h2><ul>", body)

		for i := range gelieferteArtikel {
			body = fmt.Sprintf("%s<li><b>%s</b> - %s</li>", body, gelieferteArtikel[i].Artikelnummer, gelieferteArtikel[i].Name)
		}
		body = fmt.Sprintf("%s</ul>", body)
	}

	if len(neuePreise) > 0 {
		body = fmt.Sprintf("%s<br><br><h2>Preisänderungen</h2><ul>", body)

		for i := range neuePreise {
			var alterPreis float64
			var neuerPreis float64
			var err error

			alterPreis, err = strconv.ParseFloat(neuePreise[i].Alterpreis.String, 64)
			if err != nil {
				panic(err)
			}
			neuerPreis, err = strconv.ParseFloat(neuePreise[i].Neuerpreis.String, 64)
			if err != nil {
				panic(err)
			}

			if neuerPreis != alterPreis {
				body = fmt.Sprintf("%s<li><b>%s</b> - %s: %.2f ➡️ %.2f ", body, neuePreise[i].Artikelnummer, neuePreise[i].Name, alterPreis, neuerPreis)
				var altFloat float64
				var neuFloat float64
				if neuePreise[i].Alterpreis.Valid {
					altFloat, _ = strconv.ParseFloat(neuePreise[i].Alterpreis.String, 64)
				}
				if neuePreise[i].Neuerpreis.Valid {
					neuFloat, _ = strconv.ParseFloat(neuePreise[i].Neuerpreis.String, 64)
				}
				absolute := neuFloat - altFloat
				prozent := ((altFloat / altFloat) * 100) - 100
				body = fmt.Sprintf("%s(%.2f %% // %.2f €)</li>", body, prozent, absolute)
			}

		}
		body = fmt.Sprintf("%s</ul>", body)
	}

	body = fmt.Sprintf("%s<h2>Aktuelle Lagerwerte</h2><p><b>Lagerwert Verfügbare Artikel:</b> %.2f €</p><p><b>Lagerwert alle lagernde Artikel:</b> %.2f €</p>", body, wertVerfügbar, wertBestand)
	body = fmt.Sprintf("%s<p>Wert in aktuellen Aufträgen: %.2f €", body, wertBestand-wertVerfügbar)

	if len(SN) > 0 {
		body = fmt.Sprintf("%s<h2>Artikel mit alten Seriennummern</h2><p>Nachfolgende Artikel sollten mit erhöhter Prioriät verkauf werden, da die Seriennummern bereits sehr alt sind. Gegebenenfalls sind die Artikel bereits außerhalb der Herstellergarantie!</p>", body)
		body = fmt.Sprintf("%s<p>Folgende Werte gelten:</p>", body)
		body = fmt.Sprintf("%s<p>Wortmann: Angebene Garantielaufzeit + 2 Monate ab Kaufdatum CompEx</p>", body)
		body = fmt.Sprintf("%s<p>Lenovo: Angegebene Garantielaufzeit ab Kauf CompEx</p>", body)
		body = fmt.Sprintf("%s<p>Bei allen anderen Herstellern gilt teilweise das Kaufdatum des Kunden. <br>Falls sich dies ändern sollte, wird es in der Aufzählung ergänzt.</p>", body)

		body = fmt.Sprintf("%s<p>Erklärungen der Farben:</p>", body)
		body = fmt.Sprintf("%s<p><span style='background-color: \"#f45865\"'>ROT:</span> Artikel ist bereits seit mehr als 2 Jahren lagernd und sollte schnellstens Verkauft werden!</p>", body)
		body = fmt.Sprintf("%s<p><span style='background-color: \"#fff200\"'>Gelb:</span> Artikel ist bereits seit mehr als 1 Jahr lagernd!</p>", body)

		body = fmt.Sprintf("%s<table><thead>", body)
		body = fmt.Sprintf("%s<tr>", body)
		body = fmt.Sprintf("%s<th>Artikelnummer</th>", body)
		body = fmt.Sprintf("%s<th>Name</th>", body)
		body = fmt.Sprintf("%s<th>Bestand</th>", body)
		body = fmt.Sprintf("%s<th>Verfügbar</th>", body)
		body = fmt.Sprintf("%s<th>Garantiebeginn des ältesten Artikels</th>", body)
		body = fmt.Sprintf("%s</tr>", body)
		body = fmt.Sprintf("%s</thead>", body)
		body = fmt.Sprintf("%s</thbody>", body)
		for i := range SN {
			year, _, _ := time.Now().Date()
			tmp := strings.Split(strings.Replace(strings.Split(SN[i].GeBeginn, "T")[0], "-", ".", -1), ".")
			year_tmp, err := strconv.Atoi(tmp[0])
			if err != nil {
				log.Fatal("SendMail: Fehler beim voncertieren von string zu int (year) in GetAlteSeriennummern!", err)
			}

			GarantieBeginn := fmt.Sprintf("%s.%s.%s", tmp[2], tmp[1], tmp[0])
			diff := year - year_tmp
			if diff >= 2 {
				body = fmt.Sprintf("%s<tr style='background-color: \"#f45865\"'>", body)
			} else if diff >= 1 {
				body = fmt.Sprintf("%s<tr style='background-color: \"#fff200\"'>", body)
			} else {
				body = fmt.Sprintf("%s<tr>", body)
			}
			body = fmt.Sprintf("%s<td>%s</td>", body, SN[i].ArtNr)
			body = fmt.Sprintf("%s<td>%s</td>", body, SN[i].Suchbegriff)
			body = fmt.Sprintf("%s<td>%v</td>", body, SN[i].Bestand)
			body = fmt.Sprintf("%s<td>%v</td>", body, SN[i].Verfügbar)
			body = fmt.Sprintf("%s<td>%s</td>", body, GarantieBeginn)
			body = fmt.Sprintf("%s</tr>", body)

		}
		body = fmt.Sprintf("%s</tbody></table>", body)
	}

	if len(teureArtikel) > 0 {
		body = fmt.Sprintf("%s<h2>Top 10: Die teuersten Artikel inkl. aktive Aufträge</h2><table><thead><tr><th>Artikelnummer</th><th>Name</th><th>Bestand</th><th>Einzelpreis</th><th>Summe</th></tr></thead><tbody>", body)

		for i := range teureArtikel {
			body = fmt.Sprintf("%s<tr><td>%s</td><td>%s</td><td>%d</td><td>%.2f €</td><td>%.2f €</td></tr>", body, teureArtikel[i].Artikelnummer, teureArtikel[i].Artikelname, teureArtikel[i].Bestand, teureArtikel[i].EK, teureArtikel[i].Summe)
		}
		body = fmt.Sprintf("%s</tbody></table>", body)
	}

	if len(teureVerfArtikel) > 0 {
		body = fmt.Sprintf("%s<h2>Top 10: Die teuersten Artikel exkl. aktive Aufträge</h2><table><thead><tr><th>Artikelnummer</th><th>Name</th><th>Bestand</th><th>Einzelpreis</th><th>Summe</th></tr></thead><tbody>", body)

		for i := range teureVerfArtikel {
			body = fmt.Sprintf("%s<tr><td>%s</td><td>%s</td><td>%d</td><td>%.2f €</td><td>%.2f €</td></tr>", body, teureVerfArtikel[i].Artikelnummer, teureVerfArtikel[i].Artikelname, teureVerfArtikel[i].Bestand, teureVerfArtikel[i].EK, teureVerfArtikel[i].Summe)

		}
		body = fmt.Sprintf("%s</tbody></table>", body)
	}

	if len(leichen) > 0 {
		body = fmt.Sprintf("%s<h2>Top 20: Leichen bei CE</h2><table><thead><tr><th>Artikelnummer</th><th>Name</th><th>Bestand</th><th>Verfügbar</th><th>Letzter Umsatz:</th><th>Wert im Lager:</th></tr></thead><tbody>", body)
		for i := range leichen {
			summe := float64(leichen[i].Verfügbar) * leichen[i].EK
			var LetzterUmsatz string
			if leichen[i].LetzterUmsatz == "1899-12-30T00:00:00Z" {
				LetzterUmsatz = "nie"
			} else {
				tmp := strings.Split(strings.Replace(strings.Split(leichen[i].LetzterUmsatz, "T")[0], "-", ".", -1), ".")
				LetzterUmsatz = fmt.Sprintf("%s.%s.%s", tmp[2], tmp[1], tmp[0])
			}
			bestand := leichen[i].Bestand
			verf := leichen[i].Verfügbar
			artNr := leichen[i].Artikelnummer
			name := leichen[i].Artikelname
			body = fmt.Sprintf("%s<tr><td>%s</td><td>%s</td><td>%d</td><td>%d</td><td>%s</td><td>%.2f€</td></tr>", body, artNr, name, bestand, verf, LetzterUmsatz, summe)
		}
		body = fmt.Sprintf("%s</tbody></table>", body)
	}

	d := gomail.NewDialer(MAIL_SERVER, MAIL_PORT, MAIL_USER, MAIL_PASSWORD)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	s, err := d.Dial()
	if err != nil {
		fehler := err.Error()
		return fehler
	}
	m := gomail.NewMessage()

	for i := range Mitarbeiter {
		if Mitarbeiter[i].Email.Valid && len(Mitarbeiter[i].Email.String) > 1 {

			// Set Mail Sender
			m.SetHeader("From", MAIL_FROM)
			// Receiver
			m.SetHeader("To", Mitarbeiter[i].Email.String)
			// Set Subject
			m.SetHeader("Subject", fmt.Sprintf("Warenlieferung vom %v", time.Now().Format(time.DateOnly)))
			// Set Body
			m.SetBody("text/html", body)

			if err := gomail.Send(s, m); err != nil {
				fmt.Println(err)
				fehler := err.Error()

				return fehler
			}

			m.Reset()
		}
	}

	return "OK"
}

func (a *App) SendInfo(Auftrag string, Mail string) string {
	env := config.GetEnv()
	// Get Mail Settings
	MAIL_FROM := env.MAIL_FROM
	MAIL_SERVER := env.MAIL_SERVER
	MAIL_PORT := env.MAIL_PORT
	MAIL_USER := env.MAIL_USER
	MAIL_PASSWORD := env.MAIL_PASSWORD

	// Get and Parse HTML Template
	t := template.New("mail.html")

	t, err := t.ParseFiles("mail.html")
	if err != nil {
		log.Println(err)
	}

	var tpl bytes.Buffer
	if err := t.Execute(&tpl, nil); err != nil {
		log.Println(err)
	}

	result := tpl.String()

	// Create Mail
	m := gomail.NewMessage()

	// Set Mail Sender
	m.SetHeader("From", MAIL_FROM)
	// Receiver
	m.SetHeader("To", Mail)
	// BCC
	m.SetHeader("Bcc", "service@computer-extra.de")
	// Set Subject
	m.SetHeader("Subject", fmt.Sprintf("Ihre Bestellung %s", Auftrag))
	// Set Body
	m.SetBody("text/html", result)

	d := gomail.NewDialer(MAIL_SERVER, MAIL_PORT, MAIL_USER, MAIL_PASSWORD)

	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	if err := d.DialAndSend(m); err != nil {

		fehler := err.Error()

		return fehler
	}

	return "OK"
}

func getSageConnectionString() string {
	env := config.GetEnv()
	server := env.SAGE_SERVER
	db := env.SAGE_DB
	user := env.SAGE_USER
	password := env.SAGE_PASS
	port := env.SAGE_PORT

	return fmt.Sprintf("server=%s;database=%s;user id=%s;password=%s;port=%d", server, db, user, password, port)
}

func syncAussteller(ctx context.Context) error {

	// BUG: Geht nicht.
	// FIX: Complete rewrite!
	env := config.GetEnv()

	conn, err := sql.Open("sqlserver", getSageConnectionString())
	if err != nil {
		return err
	}
	defer conn.Close()
	sage_query := "select sg_auf_artikel.SG_AUF_ARTIKEL_PK, sg_auf_artikel.ARTNR, sg_auf_artikel.SUCHBEGRIFF, sg_auf_artikel.ZUSTEXT1, sg_auf_vkpreis.PR01 FROM sg_auf_artikel INNER JOIN sg_auf_vkpreis ON sg_auf_artikel.SG_AUF_ARTIKEL_PK = sg_auf_vkpreis.SG_AUF_ARTIKEL_FK"
	rows, err := conn.Query(sage_query)
	if err != nil {
		return err
	}
	defer rows.Close()

	var Sage []AusstellerArtikel
	for rows.Next() {
		var Id sql.NullInt64
		var Artikelnummer sql.NullString
		var Artikelname sql.NullString
		var Specs sql.NullString
		var Preis sql.NullFloat64

		if err := rows.Scan(&Id, &Artikelnummer, &Artikelname, &Specs, &Preis); err != nil {
			return err
		}
		if Id.Valid && Artikelnummer.Valid && Artikelname.Valid && Specs.Valid && Preis.Valid {
			var tmp AusstellerArtikel
			tmp.Id = int(Id.Int64)
			tmp.Artikelnummer = Artikelnummer.String
			tmp.Artikelname = Artikelname.String
			tmp.Preis = Preis.Float64
			tmp.Specs = Specs.String
			Sage = append(Sage, tmp)
		}
	}

	datebase, err := sql.Open("mysql", env.DATABASE_URL)
	if err != nil {
		return err
	}
	datebase.SetConnMaxIdleTime(time.Minute * 3)
	datebase.SetMaxOpenConns(10)
	datebase.SetMaxIdleConns(10)
	queries := db.New(datebase)

	if len(Sage) > 0 {
		for i := range Sage {
			id := Sage[i].Id
			nummer := Sage[i].Artikelnummer
			name := strings.ReplaceAll(Sage[i].Artikelname, "'", "\"")
			spec := strings.ReplaceAll(Sage[i].Specs, "'", "\"")
			price := Sage[i].Preis
			_, err := queries.InsertAussteller(ctx, db.InsertAusstellerParams{
				ID:            int32(id),
				Artikelnummer: nummer,
				Artikelname:   name,
				Specs:         spec,
				Preis:         fmt.Sprintf("%.2f", price),
				Artikelname_2: name,
				Specs_2:       spec,
				Preis_2:       fmt.Sprintf("%.2f", price),
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func readSage() ([]SageArtikel, error) {
	env := config.GetEnv()

	connString := fmt.Sprintf("server=%s;database=%s;user id=%s;password=%s;port=%d", env.SAGE_SERVER, env.SAGE_DB, env.SAGE_USER, env.SAGE_PASS, env.SAGE_PORT)

	conn, err := sql.Open("sqlserver", connString)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	rows, err := conn.Query("SELECT sg_auf_artikel.SG_AUF_ARTIKEL_PK, sg_auf_artikel.ARTNR, sg_auf_artikel.SUCHBEGRIFF, sg_auf_vkpreis.PR01 FROM sg_auf_artikel INNER JOIN sg_auf_vkpreis ON (sg_auf_artikel.SG_AUF_ARTIKEL_PK = sg_auf_vkpreis.SG_AUF_ARTIKEL_FK)")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var artikel []SageArtikel

	for rows.Next() {
		var art SageArtikel
		var Artikelnummer sql.NullString
		var Suchbegriff sql.NullString
		var Price sql.NullFloat64

		if err := rows.Scan(&art.Id, &Artikelnummer, &Suchbegriff, &Price); err != nil {
			return nil, err
		}
		if Artikelnummer.Valid && Suchbegriff.Valid && Price.Valid {
			art.Artikelnummer = Artikelnummer.String
			art.Suchbegriff = Suchbegriff.String
			art.Preis = Price.Float64
			artikel = append(artikel, art)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return artikel, nil
}

func readAccessDb() ([]AccessArtikel, error) {
	env := config.GetEnv()

	conn, err := sql.Open("adodb", fmt.Sprintf("Provider=Microsoft.ACE.OLEDB.12.0;Data Source=%s;", env.ACCESS_DB))
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	rows, err := conn.Query("SELECT ID, Artikelnummer, Artikeltext, Preis FROM Artikel")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var artikel []AccessArtikel

	for rows.Next() {
		var art AccessArtikel
		if err := rows.Scan(&art.Id, &art.Artikelnummer, &art.Artikeltext, &art.Preis); err != nil {
			return nil, err
		}
		artikel = append(artikel, art)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return artikel, nil
}

func syncDb(sage []SageArtikel, label []AccessArtikel) error {
	var updates []AccessArtikel
	var create []AccessArtikel

	for i := range sage {
		var found bool
		found = false
		for x := range label {
			if sage[i].Id == label[x].Id {
				found = true
				break
			}
		}
		var art AccessArtikel
		art.Id = sage[i].Id
		art.Artikelnummer = sage[i].Artikelnummer
		art.Preis = sage[i].Preis
		art.Artikeltext = sage[i].Suchbegriff
		if found {
			updates = append(updates, art)
		} else {
			create = append(create, art)
		}
	}

	err := insert(create)
	if err != nil {
		return err
	}
	err = update(updates)
	if err != nil {
		return err
	}
	return nil
}

func insert(create []AccessArtikel) error {
	env := config.GetEnv()

	conn, err := sql.Open("adodb", fmt.Sprintf("Provider=Microsoft.ACE.OLEDB.12.0;Data Source=%s;", env.ACCESS_DB))
	if err != nil {
		return err
	}
	defer conn.Close()

	stmt, err := conn.Prepare("INSERT INTO Artikel (ID, Artikelnummer, Artikeltext, Preis) VALUES (?,?,?,?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for x := range create {
		if _, err := stmt.Exec(create[x].Id, create[x].Artikelnummer, strings.ReplaceAll(create[x].Artikeltext, "'", "\""), create[x].Preis); err != nil {
			return err
		}
	}
	return nil
}

func update(update []AccessArtikel) error {
	env := config.GetEnv()

	conn, err := sql.Open("adodb", fmt.Sprintf("Provider=Microsoft.ACE.OLEDB.12.0;Data Source=%s;", env.ACCESS_DB))
	if err != nil {
		return err
	}
	defer conn.Close()

	stmt, err := conn.Prepare("UPDATE Artikel SET Artikelnummer=?, Artikeltext=?, Preis=? where ID=?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for x := range update {
		if _, err := stmt.Exec(update[x].Artikelnummer, strings.ReplaceAll(update[x].Artikeltext, "'", "\""), update[x].Preis, update[x].Id); err != nil {
			return err
		}
	}
	return nil
}

func sortProducts(Products []db.Warenlieferung) ([]db.Warenlieferung, []db.Warenlieferung, []db.Warenlieferung, error) {
	Sage, err := getAllProductsFromSage()
	if err != nil {
		return nil, nil, nil, err
	}
	History, err := getLagerHistory()
	if err != nil {
		return nil, nil, nil, err
	}
	Prices, err := getPrices()
	if err != nil {
		return nil, nil, nil, err
	}

	var neueArtikel []db.Warenlieferung
	var gelieferteArtikel []db.Warenlieferung
	var geliefert []int
	var neuePreise []db.Warenlieferung

	if len(Products) <= 0 {
		for i := range Sage {
			neu := db.Warenlieferung{
				ID:            int32(Sage[i].Id),
				Artikelnummer: Sage[i].Artikelnummer,
				Name:          Sage[i].Suchbegriff,
			}
			neueArtikel = append(neueArtikel, neu)
		}
	} else {
		for i := range History {
			if History[i].Action == "Insert" {
				geliefert = append(geliefert, History[i].Id)
			}
		}
		for i := range Sage {
			var found bool
			found = false
			for y := 0; y < len(geliefert); y++ {
				if Sage[i].Id == geliefert[y] {
					prod := db.Warenlieferung{
						ID:   int32(Sage[i].Id),
						Name: Sage[i].Suchbegriff,
					}
					gelieferteArtikel = append(gelieferteArtikel, prod)
				}
			}
			for x := 0; x < len(Products); x++ {
				if Sage[i].Id == int(Products[x].ID) {
					found = true
					break
				}
			}
			if !found {
				neu := db.Warenlieferung{
					ID:            int32(Sage[i].Id),
					Artikelnummer: Sage[i].Artikelnummer,
					Name:          Sage[i].Suchbegriff,
				}
				neueArtikel = append(neueArtikel, neu)
			}
		}
		for i := range Prices {
			var temp db.Warenlieferung
			var found bool
			idx := 0
			if len(neuePreise) > 0 {
				for x := 0; x < len(neuePreise); x++ {
					if neuePreise[x].ID == int32(Prices[i].Id) {
						found = true
						temp = neuePreise[x]
						idx = x
					}
				}
			}
			if !found {
				temp.ID = int32(Prices[i].Id)
				temp.Preis = sql.NullTime{Time: time.Now(), Valid: true}
			}
			if Prices[i].Action == "Insert" {
				temp.Neuerpreis = sql.NullString{String: fmt.Sprintf("%.2f", Prices[i].Price), Valid: true}
			}
			if Prices[i].Action == "Delete" {
				temp.Alterpreis = sql.NullString{String: fmt.Sprintf("%.2f", Prices[i].Price), Valid: true}
			}
			if idx > 0 {
				var altFloat float64
				var neuFloat float64
				if temp.Alterpreis.Valid {
					altFloat, _ = strconv.ParseFloat(temp.Alterpreis.String, 32)
				}
				if altFloat > 0 {
					neuePreise[idx].Alterpreis = sql.NullString{String: temp.Alterpreis.String, Valid: true}
				}
				if temp.Neuerpreis.Valid {
					neuFloat, _ = strconv.ParseFloat(temp.Neuerpreis.String, 32)
				}
				if neuFloat > 0 {
					neuePreise[idx].Neuerpreis = sql.NullString{String: temp.Neuerpreis.String, Valid: true}
				}
			} else {
				neuePreise = append(neuePreise, temp)
			}
		}
	}

	return neueArtikel, gelieferteArtikel, neuePreise, nil
}

func getAllProductsFromSage() ([]Artikel, error) {
	flag.Parse()

	var artikel []Artikel

	connString := getSageConnectionString()

	conn, err := sql.Open("sqlserver", connString)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	rows, err := conn.Query("SELECT SG_AUF_ARTIKEL_PK, ARTNR, SUCHBEGRIFF FROM sg_auf_artikel")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var art Artikel
		var Artikelnummer sql.NullString
		var Suchbegriff sql.NullString

		if err := rows.Scan(&art.Id, &Artikelnummer, &Suchbegriff); err != nil {
			return nil,
				err
		}
		if Artikelnummer.Valid {
			art.Artikelnummer = Artikelnummer.String
		}
		if Suchbegriff.Valid {
			art.Suchbegriff = Suchbegriff.String
		}
		if Suchbegriff.Valid && Artikelnummer.Valid {
			artikel = append(artikel, art)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return artikel, nil
}

func getLagerHistory() ([]History, error) {
	var history []History

	queryString := fmt.Sprintf("SELECT SG_AUF_ARTIKEL_FK, Hist_Action FROM sg_auf_lager_history WHERE BEWEGUNG >= 0 AND BEMERKUNG LIKE 'Warenlieferung:%%' AND convert(varchar, Hist_Datetime, 105) = convert(varchar, getdate(), 105)")

	connString := getSageConnectionString()

	conn, err := sql.Open("sqlserver", connString)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	rows, err := conn.Query(queryString)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var hist History
		var Action sql.NullString

		if err := rows.Scan(&hist.Id, &Action); err != nil {
			return nil, err
		}
		if Action.Valid {
			hist.Action = Action.String
			history = append(history, hist)
		}
	}

	return history, nil
}

func getPrices() ([]Price, error) {
	var prices []Price

	queryString := "SELECT Hist_Action, SG_AUF_ARTIKEL_FK, PR01 FROM sg_auf_vkpreis_history WHERE convert(varchar, Hist_Datetime, 105) = convert(varchar, getdate(), 105)"

	connString := getSageConnectionString()

	conn, err := sql.Open("sqlserver", connString)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	rows, err := conn.Query(queryString)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var price Price
		var Action sql.NullString
		var p sql.NullFloat64

		if err := rows.Scan(&Action, &price.Id, &p); err != nil {
			return nil, fmt.Errorf("GetPrices: Row Error: %s", err)
		}
		if p.Valid {
			price.Price = float32(p.Float64)
		}
		if Action.Valid {
			price.Action = Action.String
		}
		if p.Valid && Action.Valid {
			prices = append(prices, price)
		}
	}

	return prices, nil
}

func getLeichen() ([]Leichen, error) {
	var artikel []Leichen

	conn, err := sql.Open("sqlserver", getSageConnectionString())
	if err != nil {
		return nil, err
	}

	defer conn.Close()
	rows, err := conn.Query("SELECT TOP 20 ARTNR, SUCHBEGRIFF, BESTAND, VERFUEGBAR, LetzterUmsatz, EKPR01 FROM sg_auf_artikel WHERE VERFUEGBAR > 0 ORDER BY LetzterUmsatz ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var art Leichen
		var Artikelnummer sql.NullString
		var Artikelname sql.NullString
		var Bestand sql.NullInt16
		var Verfügbar sql.NullInt16
		var EK sql.NullFloat64
		var LetzerUmsatz sql.NullString

		if err := rows.Scan(&Artikelnummer, &Artikelname, &Bestand, &Verfügbar, &LetzerUmsatz, &EK); err != nil {
			return nil,
				err
		}
		if Artikelnummer.Valid {
			art.Artikelnummer = Artikelnummer.String
		}
		if Artikelname.Valid {
			art.Artikelname = Artikelname.String
		}
		if Bestand.Valid {
			art.Bestand = Bestand.Int16
		}
		if EK.Valid {
			art.EK = EK.Float64
		}
		if Verfügbar.Valid {
			art.Verfügbar = Verfügbar.Int16
		}
		if LetzerUmsatz.Valid {
			art.LetzterUmsatz = LetzerUmsatz.String
		}

		if Artikelnummer.Valid && Artikelname.Valid && Bestand.Valid && EK.Valid && LetzerUmsatz.Valid && Verfügbar.Valid {
			artikel = append(artikel, art)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return artikel, nil
}

func getHighestVerfSum() ([]VerfArtikel, error) {
	var artikel []VerfArtikel
	database, err := sql.Open("sqlserver", getSageConnectionString())
	if err != nil {
		return nil, err
	}
	defer database.Close()
	rows, err := database.Query("SELECT TOP 10 ARTNR, SUCHBEGRIFF, BESTAND, VERFUEGBAR, EKPR01, VERFUEGBAR * EKPR01 as Summe FROM sg_auf_artikel WHERE VERFUEGBAR > 0 ORDER BY Summe DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var art VerfArtikel
		var Artikelnummer sql.NullString
		var Artikelname sql.NullString
		var Bestand sql.NullInt16
		var Verfügbar sql.NullInt16
		var EK sql.NullFloat64
		var Summe sql.NullFloat64

		if err := rows.Scan(&Artikelnummer, &Artikelname, &Bestand, &Verfügbar, &EK, &Summe); err != nil {
			return nil,
				err
		}
		if Artikelnummer.Valid {
			art.Artikelnummer = Artikelnummer.String
		}
		if Artikelname.Valid {
			art.Artikelname = Artikelname.String
		}
		if Bestand.Valid {
			art.Bestand = Bestand.Int16
		}
		if EK.Valid {
			art.EK = EK.Float64
		}
		if Summe.Valid {
			art.Summe = Summe.Float64
		}
		if Verfügbar.Valid {
			art.Verfügbar = Verfügbar.Int16
		}

		if Artikelnummer.Valid && Artikelname.Valid && Bestand.Valid && EK.Valid && Summe.Valid && Verfügbar.Valid {
			artikel = append(artikel, art)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return artikel, nil
}

func getHighestSum() ([]SummenArtikel, error) {
	var artikel []SummenArtikel

	conn, err := sql.Open("sqlserver", getSageConnectionString())
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	rows, err := conn.Query("SELECT TOP 10 ARTNR, SUCHBEGRIFF, BESTAND, EKPR01, BESTAND * EKPR01 as Summe FROM sg_auf_artikel WHERE BESTAND > 0 ORDER BY Summe DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var art SummenArtikel
		var Artikelnummer sql.NullString
		var Artikelname sql.NullString
		var Bestand sql.NullInt16
		var EK sql.NullFloat64
		var Summe sql.NullFloat64

		if err := rows.Scan(&Artikelnummer, &Artikelname, &Bestand, &EK, &Summe); err != nil {
			return nil,
				err
		}
		if Artikelnummer.Valid {
			art.Artikelnummer = Artikelnummer.String
		}
		if Artikelname.Valid {
			art.Artikelname = Artikelname.String
		}
		if Bestand.Valid {
			art.Bestand = Bestand.Int16
		}
		if EK.Valid {
			art.EK = EK.Float64
		}
		if Summe.Valid {
			art.Summe = Summe.Float64
		}

		if Artikelnummer.Valid && Artikelname.Valid && Bestand.Valid && EK.Valid && Summe.Valid {
			artikel = append(artikel, art)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return artikel, nil
}

func getLagerWert() (float64, float64, error) {
	var wertBestand float64
	var wertVerfügbar float64
	wertBestand = 0
	wertVerfügbar = 0

	conn, err := sql.Open("sqlserver", getSageConnectionString())
	if err != nil {
		return 0, 0, err
	}
	defer conn.Close()
	rows, err := conn.Query("SELECT BESTAND, VERFUEGBAR, EKPR01 FROM sg_auf_artikel WHERE BESTAND > 0")
	if err != nil {
		return 0, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var Bestand sql.NullInt16
		var Verfügbar sql.NullInt16
		var Ek sql.NullFloat64

		if err := rows.Scan(&Bestand, &Verfügbar, &Ek); err != nil {
			return 0, 0,
				err
		}
		if Bestand.Valid && Ek.Valid {
			wertBestand = wertBestand + (float64(Bestand.Int16) * Ek.Float64)
		}
		if Verfügbar.Valid && Ek.Valid {
			wertVerfügbar = wertVerfügbar + (float64(Verfügbar.Int16) * Ek.Float64)
		}
	}
	if err := rows.Err(); err != nil {
		return 0, 0, err
	}

	return wertBestand, wertVerfügbar, nil
}

func getAlteSeriennummern() ([]AlteSeriennummer, error) {
	var artikel []AlteSeriennummer

	conn, err := sql.Open("sqlserver", getSageConnectionString())
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	rows, err := conn.Query("SELECT sg_auf_artikel.ARTNR, sg_auf_artikel.SUCHBEGRIFF, sg_auf_artikel.BESTAND, sg_auf_artikel.VERFUEGBAR, sg_auf_snr.GE_Beginn FROM sg_auf_artikel INNER JOIN sg_auf_snr ON sg_auf_artikel.SG_AUF_ARTIKEL_PK = sg_auf_snr.SG_AUF_ARTIKEL_FK  WHERE sg_auf_artikel.VERFUEGBAR > 0 AND sg_auf_snr.SNR_STATUS != 2 AND sg_auf_snr.GE_Beginn <= DATEADD(month, DATEDIFF(month, 0, DATEADD(MONTH,-1,GETDATE())), 0) ORDER BY sg_auf_snr.GE_Beginn ")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var art AlteSeriennummer
		var Artikelnummer sql.NullString
		var Suchbegriff sql.NullString
		var Bestand sql.NullInt16
		var Verfügbar sql.NullInt16
		var Garantie sql.NullString

		if err := rows.Scan(&Artikelnummer, &Suchbegriff, &Bestand, &Verfügbar, &Garantie); err != nil {
			return nil,
				err
		}

		if Artikelnummer.Valid && Suchbegriff.Valid && Bestand.Valid && Verfügbar.Valid && Garantie.Valid {
			art.ArtNr = Artikelnummer.String
			art.Suchbegriff = Suchbegriff.String
			art.Bestand = int(Bestand.Int16)
			art.Verfügbar = int(Verfügbar.Int16)
			art.GeBeginn = Garantie.String
			if !slices.Contains(artikel, art) {
				artikel = append(artikel, art)
			}
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return artikel, nil
}
