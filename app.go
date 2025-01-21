package main

import (
	"ServiceApp/config"
	"ServiceApp/db"
	"context"
	"crypto/tls"
	"database/sql"
	"flag"
	"fmt"
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
	err := syncAussteller()
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

	body := `
		<html xmlns=http://www.w3.org/TR/REC-html40 xmlns:m=http://schemas.microsoft.com/office/2004/12/omml xmlns:o=urn:schemas-microsoft-com:office:office xmlns:v=urn:schemas-microsoft-com:vml xmlns:w=urn:schemas-microsoft-com:office:word><meta content="text/html; charset=windows-1252"http-equiv=Content-Type><meta content=Word.Document name=ProgId><meta content="Microsoft Word 15"name=Generator><meta content="Microsoft Word 15"name=Originator><link href=Info-Dateien/filelist.xml rel=File-List><!--[if gte mso 9
		]><xml><o:documentproperties><o:author>Verkauf (Computer Extra KG)</o:author><o:template>Normal</o:template><o:lastauthor>Julian Thurian (Computer Extra)</o:lastauthor><o:revision>2</o:revision><o:created>2024-04-30T10:22:00Z</o:created><o:lastsaved>2024-04-30T10:22:00Z</o:lastsaved><o:pages>1</o:pages><o:words>146</o:words><o:characters>1269</o:characters><o:lines>10</o:lines><o:paragraphs>2</o:paragraphs><o:characterswithspaces>1413</o:characterswithspaces><o:version>16.00</o:version></o:documentproperties><o:officedocumentsettings><o:allowpng></o:officedocumentsettings></xml><![endif]--><link href=Info-Dateien/themedata.thmx rel=themeData><link href=Info-Dateien/colorschememapping.xml rel=colorSchemeMapping><!--[if gte mso 9
		]><xml><w:worddocument><w:spellingstate>Clean</w:spellingstate><w:grammarstate>Clean</w:grammarstate><w:trackmoves><w:trackformatting><w:hyphenationzone>21</w:hyphenationzone><w:validateagainstschemas><w:saveifxmlinvalid>false</w:saveifxmlinvalid><w:ignoremixedcontent>false</w:ignoremixedcontent><w:alwaysshowplaceholdertext>false</w:alwaysshowplaceholdertext><w:donotpromoteqf><w:lidthemeother>DE</w:lidthemeother><w:lidthemeasian>X-NONE</w:lidthemeasian><w:lidthemecomplexscript>X-NONE</w:lidthemecomplexscript><w:compatibility><w:breakwrappedtables><w:useword2010tablestylerules><w:splitpgbreakandparamark></w:compatibility><m:mathpr><m:mathfont m:val="Cambria Math"><m:brkbin m:val=before><m:brkbinsub m:val=--><m:smallfrac m:val=off><m:dispdef><m:lmargin m:val=0><m:rmargin m:val=0><m:defjc m:val=centerGroup><m:wrapindent m:val=1440><m:intlim m:val=subSup><m:narylim m:val=undOvr></m:mathpr></w:worddocument></xml><![endif]--><!--[if gte mso 9
		]><xml><w:latentstyles deflockedstate=false defpriority=99 defqformat=false defsemihidden=false defunhidewhenused=false latentstylecount=376><w:lsdexception locked=false name=Normal priority=0 qformat=true><w:lsdexception locked=false name="heading 1"priority=9 qformat=true><w:lsdexception locked=false name="heading 2"priority=9 qformat=true semihidden=true unhidewhenused=true><w:lsdexception locked=false name="heading 3"priority=9 qformat=true semihidden=true unhidewhenused=true><w:lsdexception locked=false name="heading 4"priority=9 qformat=true semihidden=true unhidewhenused=true><w:lsdexception locked=false name="heading 5"priority=9 qformat=true semihidden=true unhidewhenused=true><w:lsdexception locked=false name="heading 6"priority=9 qformat=true semihidden=true unhidewhenused=true><w:lsdexception locked=false name="heading 7"priority=9 qformat=true semihidden=true unhidewhenused=true><w:lsdexception locked=false name="heading 8"priority=9 qformat=true semihidden=true unhidewhenused=true><w:lsdexception locked=false name="heading 9"priority=9 qformat=true semihidden=true unhidewhenused=true><w:lsdexception locked=false name="index 1"semihidden=true unhidewhenused=true><w:lsdexception locked=false name="index 2"semihidden=true unhidewhenused=true><w:lsdexception locked=false name="index 3"semihidden=true unhidewhenused=true><w:lsdexception locked=false name="index 4"semihidden=true unhidewhenused=true><w:lsdexception locked=false name="index 5"semihidden=true unhidewhenused=true><w:lsdexception locked=false name="index 6"semihidden=true unhidewhenused=true><w:lsdexception locked=false name="index 7"semihidden=true unhidewhenused=true><w:lsdexception locked=false name="index 8"semihidden=true unhidewhenused=true><w:lsdexception locked=false name="index 9"semihidden=true unhidewhenused=true><w:lsdexception locked=false name="toc 1"priority=39 semihidden=true unhidewhenused=true><w:lsdexception locked=false name="toc 2"priority=39 semihidden=true unhidewhenused=true><w:lsdexception locked=false name="toc 3"priority=39 semihidden=true unhidewhenused=true><w:lsdexception locked=false name="toc 4"priority=39 semihidden=true unhidewhenused=true><w:lsdexception locked=false name="toc 5"priority=39 semihidden=true unhidewhenused=true><w:lsdexception locked=false name="toc 6"priority=39 semihidden=true unhidewhenused=true><w:lsdexception locked=false name="toc 7"priority=39 semihidden=true unhidewhenused=true><w:lsdexception locked=false name="toc 8"priority=39 semihidden=true unhidewhenused=true><w:lsdexception locked=false name="toc 9"priority=39 semihidden=true unhidewhenused=true><w:lsdexception locked=false name="Normal Indent"semihidden=true unhidewhenused=true><w:lsdexception locked=false name="footnote text"semihidden=true unhidewhenused=true><w:lsdexception locked=false name="annotation text"semihidden=true unhidewhenused=true><w:lsdexception locked=false name=header semihidden=true unhidewhenused=true><w:lsdexception locked=false name=footer semihidden=true unhidewhenused=true><w:lsdexception locked=false name="index heading"semihidden=true unhidewhenused=true><w:lsdexception locked=false name=caption priority=35 qformat=true semihidden=true unhidewhenused=true><w:lsdexception locked=false name="table of figures"semihidden=true unhidewhenused=true><w:lsdexception locked=false name="envelope address"semihidden=true unhidewhenused=true><w:lsdexception locked=false name="envelope return"semihidden=true unhidewhenused=true><w:lsdexception locked=false name="footnote reference"semihidden=true unhidewhenused=true><w:lsdexception locked=false name="annotation reference"semihidden=true unhidewhenused=true><w:lsdexception locked=false name="line number"semihidden=true unhidewhenused=true><w:lsdexception locked=false name="page number"semihidden=true unhidewhenused=true><w:lsdexception locked=false name="endnote reference"semihidden=true unhidewhenused=true><w:lsdexception locked=false name="endnote text"semihidden=true unhidewhenused=true><w:lsdexception locked=false name="table of authorities"semihidden=true unhidewhenused=true><w:lsdexception locked=false name=macro semihidden=true unhidewhenused=true><w:lsdexception locked=false name="toa heading"semihidden=true unhidewhenused=true><w:lsdexception locked=false name=List semihidden=true unhidewhenused=true><w:lsdexception locked=false name="List Bullet"semihidden=true unhidewhenused=true><w:lsdexception locked=false name="List Number"semihidden=true unhidewhenused=true><w:lsdexception locked=false name="List 2"semihidden=true unhidewhenused=true><w:lsdexception locked=false name="List 3"semihidden=true unhidewhenused=true><w:lsdexception locked=false name="List 4"semihidden=true unhidewhenused=true><w:lsdexception locked=false name="List 5"semihidden=true unhidewhenused=true><w:lsdexception locked=false name="List Bullet 2"semihidden=true unhidewhenused=true><w:lsdexception locked=false name="List Bullet 3"semihidden=true unhidewhenused=true><w:lsdexception locked=false name="List Bullet 4"semihidden=true unhidewhenused=true><w:lsdexception locked=false name="List Bullet 5"semihidden=true unhidewhenused=true><w:lsdexception locked=false name="List Number 2"semihidden=true unhidewhenused=true><w:lsdexception locked=false name="List Number 3"semihidden=true unhidewhenused=true><w:lsdexception locked=false name="List Number 4"semihidden=true unhidewhenused=true><w:lsdexception locked=false name="List Number 5"semihidden=true unhidewhenused=true><w:lsdexception locked=false name=Title priority=10 qformat=true><w:lsdexception locked=false name=Closing semihidden=true unhidewhenused=true><w:lsdexception locked=false name=Signature semihidden=true unhidewhenused=true><w:lsdexception locked=false name="Default Paragraph Font"priority=1 semihidden=true unhidewhenused=true><w:lsdexception locked=false name="Body Text"semihidden=true unhidewhenused=true><w:lsdexception locked=false name="Body Text Indent"semihidden=true unhidewhenused=true><w:lsdexception locked=false name="List Continue"semihidden=true unhidewhenused=true><w:lsdexception locked=false name="List Continue 2"semihidden=true unhidewhenused=true><w:lsdexception locked=false name="List Continue 3"semihidden=true unhidewhenused=true><w:lsdexception locked=false name="List Continue 4"semihidden=true unhidewhenused=true><w:lsdexception locked=false name="List Continue 5"semihidden=true unhidewhenused=true><w:lsdexception locked=false name="Message Header"semihidden=true unhidewhenused=true><w:lsdexception locked=false name=Subtitle priority=11 qformat=true><w:lsdexception locked=false name=Salutation semihidden=true unhidewhenused=true><w:lsdexception locked=false name=Date semihidden=true unhidewhenused=true><w:lsdexception locked=false name="Body Text First Indent"semihidden=true unhidewhenused=true><w:lsdexception locked=false name="Body Text First Indent 2"semihidden=true unhidewhenused=true><w:lsdexception locked=false name="Note Heading"semihidden=true unhidewhenused=true><w:lsdexception locked=false name="Body Text 2"semihidden=true unhidewhenused=true><w:lsdexception locked=false name="Body Text 3"semihidden=true unhidewhenused=true><w:lsdexception locked=false name="Body Text Indent 2"semihidden=true unhidewhenused=true><w:lsdexception locked=false name="Body Text Indent 3"semihidden=true unhidewhenused=true><w:lsdexception locked=false name="Block Text"semihidden=true unhidewhenused=true><w:lsdexception locked=false name=Hyperlink semihidden=true unhidewhenused=true><w:lsdexception locked=false name=FollowedHyperlink semihidden=true unhidewhenused=true><w:lsdexception locked=false name=Strong priority=22 qformat=true><w:lsdexception locked=false name=Emphasis priority=20 qformat=true><w:lsdexception locked=false name="Document Map"semihidden=true unhidewhenused=true><w:lsdexception locked=false name="Plain Text"semihidden=true unhidewhenused=true><w:lsdexception locked=false name="E-mail Signature"semihidden=true unhidewhenused=true><w:lsdexception locked=false name="HTML Top of Form"semihidden=true unhidewhenused=true><w:lsdexception locked=false name="HTML Bottom of Form"semihidden=true unhidewhenused=true><w:lsdexception locked=false name="Normal (Web)"semihidden=true unhidewhenused=true><w:lsdexception locked=false name="HTML Acronym"semihidden=true unhidewhenused=true><w:lsdexception locked=false name="HTML Address"semihidden=true unhidewhenused=true><w:lsdexception locked=false name="HTML Cite"semihidden=true unhidewhenused=true><w:lsdexception locked=false name="HTML Code"semihidden=true unhidewhenused=true><w:lsdexception locked=false name="HTML Definition"semihidden=true unhidewhenused=true><w:lsdexception locked=false name="HTML Keyboard"semihidden=true unhidewhenused=true><w:lsdexception locked=false name="HTML Preformatted"semihidden=true unhidewhenused=true><w:lsdexception locked=false name="HTML Sample"semihidden=true unhidewhenused=true><w:lsdexception locked=false name="HTML Typewriter"semihidden=true unhidewhenused=true><w:lsdexception locked=false name="HTML Variable"semihidden=true unhidewhenused=true><w:lsdexception locked=false name="annotation subject"semihidden=true unhidewhenused=true><w:lsdexception locked=false name="No List"semihidden=true unhidewhenused=true><w:lsdexception locked=false name="Outline List 1"semihidden=true unhidewhenused=true><w:lsdexception locked=false name="Outline List 2"semihidden=true unhidewhenused=true><w:lsdexception locked=false name="Outline List 3"semihidden=true unhidewhenused=true><w:lsdexception locked=false name="Balloon Text"semihidden=true unhidewhenused=true><w:lsdexception locked=false name="Table Grid"priority=39><w:lsdexception locked=false name="Placeholder Text"semihidden=true><w:lsdexception locked=false name="No Spacing"priority=1 qformat=true><w:lsdexception locked=false name="Light Shading"priority=60><w:lsdexception locked=false name="Light List"priority=61><w:lsdexception locked=false name="Light Grid"priority=62><w:lsdexception locked=false name="Medium Shading 1"priority=63><w:lsdexception locked=false name="Medium Shading 2"priority=64><w:lsdexception locked=false name="Medium List 1"priority=65><w:lsdexception locked=false name="Medium List 2"priority=66><w:lsdexception locked=false name="Medium Grid 1"priority=67><w:lsdexception locked=false name="Medium Grid 2"priority=68><w:lsdexception locked=false name="Medium Grid 3"priority=69><w:lsdexception locked=false name="Dark List"priority=70><w:lsdexception locked=false name="Colorful Shading"priority=71><w:lsdexception locked=false name="Colorful List"priority=72><w:lsdexception locked=false name="Colorful Grid"priority=73><w:lsdexception locked=false name="Light Shading Accent 1"priority=60><w:lsdexception locked=false name="Light List Accent 1"priority=61><w:lsdexception locked=false name="Light Grid Accent 1"priority=62><w:lsdexception locked=false name="Medium Shading 1 Accent 1"priority=63><w:lsdexception locked=false name="Medium Shading 2 Accent 1"priority=64><w:lsdexception locked=false name="Medium List 1 Accent 1"priority=65><w:lsdexception locked=false name=Revision semihidden=true><w:lsdexception locked=false name="List Paragraph"priority=34 qformat=true><w:lsdexception locked=false name=Quote priority=29 qformat=true><w:lsdexception locked=false name="Intense Quote"priority=30 qformat=true><w:lsdexception locked=false name="Medium List 2 Accent 1"priority=66><w:lsdexception locked=false name="Medium Grid 1 Accent 1"priority=67><w:lsdexception locked=false name="Medium Grid 2 Accent 1"priority=68><w:lsdexception locked=false name="Medium Grid 3 Accent 1"priority=69><w:lsdexception locked=false name="Dark List Accent 1"priority=70><w:lsdexception locked=false name="Colorful Shading Accent 1"priority=71><w:lsdexception locked=false name="Colorful List Accent 1"priority=72><w:lsdexception locked=false name="Colorful Grid Accent 1"priority=73><w:lsdexception locked=false name="Light Shading Accent 2"priority=60><w:lsdexception locked=false name="Light List Accent 2"priority=61><w:lsdexception locked=false name="Light Grid Accent 2"priority=62><w:lsdexception locked=false name="Medium Shading 1 Accent 2"priority=63><w:lsdexception locked=false name="Medium Shading 2 Accent 2"priority=64><w:lsdexception locked=false name="Medium List 1 Accent 2"priority=65><w:lsdexception locked=false name="Medium List 2 Accent 2"priority=66><w:lsdexception locked=false name="Medium Grid 1 Accent 2"priority=67><w:lsdexception locked=false name="Medium Grid 2 Accent 2"priority=68><w:lsdexception locked=false name="Medium Grid 3 Accent 2"priority=69><w:lsdexception locked=false name="Dark List Accent 2"priority=70><w:lsdexception locked=false name="Colorful Shading Accent 2"priority=71><w:lsdexception locked=false name="Colorful List Accent 2"priority=72><w:lsdexception locked=false name="Colorful Grid Accent 2"priority=73><w:lsdexception locked=false name="Light Shading Accent 3"priority=60><w:lsdexception locked=false name="Light List Accent 3"priority=61><w:lsdexception locked=false name="Light Grid Accent 3"priority=62><w:lsdexception locked=false name="Medium Shading 1 Accent 3"priority=63><w:lsdexception locked=false name="Medium Shading 2 Accent 3"priority=64><w:lsdexception locked=false name="Medium List 1 Accent 3"priority=65><w:lsdexception locked=false name="Medium List 2 Accent 3"priority=66><w:lsdexception locked=false name="Medium Grid 1 Accent 3"priority=67><w:lsdexception locked=false name="Medium Grid 2 Accent 3"priority=68><w:lsdexception locked=false name="Medium Grid 3 Accent 3"priority=69><w:lsdexception locked=false name="Dark List Accent 3"priority=70><w:lsdexception locked=false name="Colorful Shading Accent 3"priority=71><w:lsdexception locked=false name="Colorful List Accent 3"priority=72><w:lsdexception locked=false name="Colorful Grid Accent 3"priority=73><w:lsdexception locked=false name="Light Shading Accent 4"priority=60><w:lsdexception locked=false name="Light List Accent 4"priority=61><w:lsdexception locked=false name="Light Grid Accent 4"priority=62><w:lsdexception locked=false name="Medium Shading 1 Accent 4"priority=63><w:lsdexception locked=false name="Medium Shading 2 Accent 4"priority=64><w:lsdexception locked=false name="Medium List 1 Accent 4"priority=65><w:lsdexception locked=false name="Medium List 2 Accent 4"priority=66><w:lsdexception locked=false name="Medium Grid 1 Accent 4"priority=67><w:lsdexception locked=false name="Medium Grid 2 Accent 4"priority=68><w:lsdexception locked=false name="Medium Grid 3 Accent 4"priority=69><w:lsdexception locked=false name="Dark List Accent 4"priority=70><w:lsdexception locked=false name="Colorful Shading Accent 4"priority=71><w:lsdexception locked=false name="Colorful List Accent 4"priority=72><w:lsdexception locked=false name="Colorful Grid Accent 4"priority=73><w:lsdexception locked=false name="Light Shading Accent 5"priority=60><w:lsdexception locked=false name="Light List Accent 5"priority=61><w:lsdexception locked=false name="Light Grid Accent 5"priority=62><w:lsdexception locked=false name="Medium Shading 1 Accent 5"priority=63><w:lsdexception locked=false name="Medium Shading 2 Accent 5"priority=64><w:lsdexception locked=false name="Medium List 1 Accent 5"priority=65><w:lsdexception locked=false name="Medium List 2 Accent 5"priority=66><w:lsdexception locked=false name="Medium Grid 1 Accent 5"priority=67><w:lsdexception locked=false name="Medium Grid 2 Accent 5"priority=68><w:lsdexception locked=false name="Medium Grid 3 Accent 5"priority=69><w:lsdexception locked=false name="Dark List Accent 5"priority=70><w:lsdexception locked=false name="Colorful Shading Accent 5"priority=71><w:lsdexception locked=false name="Colorful List Accent 5"priority=72><w:lsdexception locked=false name="Colorful Grid Accent 5"priority=73><w:lsdexception locked=false name="Light Shading Accent 6"priority=60><w:lsdexception locked=false name="Light List Accent 6"priority=61><w:lsdexception locked=false name="Light Grid Accent 6"priority=62><w:lsdexception locked=false name="Medium Shading 1 Accent 6"priority=63><w:lsdexception locked=false name="Medium Shading 2 Accent 6"priority=64><w:lsdexception locked=false name="Medium List 1 Accent 6"priority=65><w:lsdexception locked=false name="Medium List 2 Accent 6"priority=66><w:lsdexception locked=false name="Medium Grid 1 Accent 6"priority=67><w:lsdexception locked=false name="Medium Grid 2 Accent 6"priority=68><w:lsdexception locked=false name="Medium Grid 3 Accent 6"priority=69><w:lsdexception locked=false name="Dark List Accent 6"priority=70><w:lsdexception locked=false name="Colorful Shading Accent 6"priority=71><w:lsdexception locked=false name="Colorful List Accent 6"priority=72><w:lsdexception locked=false name="Colorful Grid Accent 6"priority=73><w:lsdexception locked=false name="Subtle Emphasis"priority=19 qformat=true><w:lsdexception locked=false name="Intense Emphasis"priority=21 qformat=true><w:lsdexception locked=false name="Subtle Reference"priority=31 qformat=true><w:lsdexception locked=false name="Intense Reference"priority=32 qformat=true><w:lsdexception locked=false name="Book Title"priority=33 qformat=true><w:lsdexception locked=false name=Bibliography priority=37 semihidden=true unhidewhenused=true><w:lsdexception locked=false name="TOC Heading"priority=39 qformat=true semihidden=true unhidewhenused=true><w:lsdexception locked=false name="Plain Table 1"priority=41><w:lsdexception locked=false name="Plain Table 2"priority=42><w:lsdexception locked=false name="Plain Table 3"priority=43><w:lsdexception locked=false name="Plain Table 4"priority=44><w:lsdexception locked=false name="Plain Table 5"priority=45><w:lsdexception locked=false name="Grid Table Light"priority=40><w:lsdexception locked=false name="Grid Table 1 Light"priority=46><w:lsdexception locked=false name="Grid Table 2"priority=47><w:lsdexception locked=false name="Grid Table 3"priority=48><w:lsdexception locked=false name="Grid Table 4"priority=49><w:lsdexception locked=false name="Grid Table 5 Dark"priority=50><w:lsdexception locked=false name="Grid Table 6 Colorful"priority=51><w:lsdexception locked=false name="Grid Table 7 Colorful"priority=52><w:lsdexception locked=false name="Grid Table 1 Light Accent 1"priority=46><w:lsdexception locked=false name="Grid Table 2 Accent 1"priority=47><w:lsdexception locked=false name="Grid Table 3 Accent 1"priority=48><w:lsdexception locked=false name="Grid Table 4 Accent 1"priority=49><w:lsdexception locked=false name="Grid Table 5 Dark Accent 1"priority=50><w:lsdexception locked=false name="Grid Table 6 Colorful Accent 1"priority=51><w:lsdexception locked=false name="Grid Table 7 Colorful Accent 1"priority=52><w:lsdexception locked=false name="Grid Table 1 Light Accent 2"priority=46><w:lsdexception locked=false name="Grid Table 2 Accent 2"priority=47><w:lsdexception locked=false name="Grid Table 3 Accent 2"priority=48><w:lsdexception locked=false name="Grid Table 4 Accent 2"priority=49><w:lsdexception locked=false name="Grid Table 5 Dark Accent 2"priority=50><w:lsdexception locked=false name="Grid Table 6 Colorful Accent 2"priority=51><w:lsdexception locked=false name="Grid Table 7 Colorful Accent 2"priority=52><w:lsdexception locked=false name="Grid Table 1 Light Accent 3"priority=46><w:lsdexception locked=false name="Grid Table 2 Accent 3"priority=47><w:lsdexception locked=false name="Grid Table 3 Accent 3"priority=48><w:lsdexception locked=false name="Grid Table 4 Accent 3"priority=49><w:lsdexception locked=false name="Grid Table 5 Dark Accent 3"priority=50><w:lsdexception locked=false name="Grid Table 6 Colorful Accent 3"priority=51><w:lsdexception locked=false name="Grid Table 7 Colorful Accent 3"priority=52><w:lsdexception locked=false name="Grid Table 1 Light Accent 4"priority=46><w:lsdexception locked=false name="Grid Table 2 Accent 4"priority=47><w:lsdexception locked=false name="Grid Table 3 Accent 4"priority=48><w:lsdexception locked=false name="Grid Table 4 Accent 4"priority=49><w:lsdexception locked=false name="Grid Table 5 Dark Accent 4"priority=50><w:lsdexception locked=false name="Grid Table 6 Colorful Accent 4"priority=51><w:lsdexception locked=false name="Grid Table 7 Colorful Accent 4"priority=52><w:lsdexception locked=false name="Grid Table 1 Light Accent 5"priority=46><w:lsdexception locked=false name="Grid Table 2 Accent 5"priority=47><w:lsdexception locked=false name="Grid Table 3 Accent 5"priority=48><w:lsdexception locked=false name="Grid Table 4 Accent 5"priority=49><w:lsdexception locked=false name="Grid Table 5 Dark Accent 5"priority=50><w:lsdexception locked=false name="Grid Table 6 Colorful Accent 5"priority=51><w:lsdexception locked=false name="Grid Table 7 Colorful Accent 5"priority=52><w:lsdexception locked=false name="Grid Table 1 Light Accent 6"priority=46><w:lsdexception locked=false name="Grid Table 2 Accent 6"priority=47><w:lsdexception locked=false name="Grid Table 3 Accent 6"priority=48><w:lsdexception locked=false name="Grid Table 4 Accent 6"priority=49><w:lsdexception locked=false name="Grid Table 5 Dark Accent 6"priority=50><w:lsdexception locked=false name="Grid Table 6 Colorful Accent 6"priority=51><w:lsdexception locked=false name="Grid Table 7 Colorful Accent 6"priority=52><w:lsdexception locked=false name="List Table 1 Light"priority=46><w:lsdexception locked=false name="List Table 2"priority=47><w:lsdexception locked=false name="List Table 3"priority=48><w:lsdexception locked=false name="List Table 4"priority=49><w:lsdexception locked=false name="List Table 5 Dark"priority=50><w:lsdexception locked=false name="List Table 6 Colorful"priority=51><w:lsdexception locked=false name="List Table 7 Colorful"priority=52><w:lsdexception locked=false name="List Table 1 Light Accent 1"priority=46><w:lsdexception locked=false name="List Table 2 Accent 1"priority=47><w:lsdexception locked=false name="List Table 3 Accent 1"priority=48><w:lsdexception locked=false name="List Table 4 Accent 1"priority=49><w:lsdexception locked=false name="List Table 5 Dark Accent 1"priority=50><w:lsdexception locked=false name="List Table 6 Colorful Accent 1"priority=51><w:lsdexception locked=false name="List Table 7 Colorful Accent 1"priority=52><w:lsdexception locked=false name="List Table 1 Light Accent 2"priority=46><w:lsdexception locked=false name="List Table 2 Accent 2"priority=47><w:lsdexception locked=false name="List Table 3 Accent 2"priority=48><w:lsdexception locked=false name="List Table 4 Accent 2"priority=49><w:lsdexception locked=false name="List Table 5 Dark Accent 2"priority=50><w:lsdexception locked=false name="List Table 6 Colorful Accent 2"priority=51><w:lsdexception locked=false name="List Table 7 Colorful Accent 2"priority=52><w:lsdexception locked=false name="List Table 1 Light Accent 3"priority=46><w:lsdexception locked=false name="List Table 2 Accent 3"priority=47><w:lsdexception locked=false name="List Table 3 Accent 3"priority=48><w:lsdexception locked=false name="List Table 4 Accent 3"priority=49><w:lsdexception locked=false name="List Table 5 Dark Accent 3"priority=50><w:lsdexception locked=false name="List Table 6 Colorful Accent 3"priority=51><w:lsdexception locked=false name="List Table 7 Colorful Accent 3"priority=52><w:lsdexception locked=false name="List Table 1 Light Accent 4"priority=46><w:lsdexception locked=false name="List Table 2 Accent 4"priority=47><w:lsdexception locked=false name="List Table 3 Accent 4"priority=48><w:lsdexception locked=false name="List Table 4 Accent 4"priority=49><w:lsdexception locked=false name="List Table 5 Dark Accent 4"priority=50><w:lsdexception locked=false name="List Table 6 Colorful Accent 4"priority=51><w:lsdexception locked=false name="List Table 7 Colorful Accent 4"priority=52><w:lsdexception locked=false name="List Table 1 Light Accent 5"priority=46><w:lsdexception locked=false name="List Table 2 Accent 5"priority=47><w:lsdexception locked=false name="List Table 3 Accent 5"priority=48><w:lsdexception locked=false name="List Table 4 Accent 5"priority=49><w:lsdexception locked=false name="List Table 5 Dark Accent 5"priority=50><w:lsdexception locked=false name="List Table 6 Colorful Accent 5"priority=51><w:lsdexception locked=false name="List Table 7 Colorful Accent 5"priority=52><w:lsdexception locked=false name="List Table 1 Light Accent 6"priority=46><w:lsdexception locked=false name="List Table 2 Accent 6"priority=47><w:lsdexception locked=false name="List Table 3 Accent 6"priority=48><w:lsdexception locked=false name="List Table 4 Accent 6"priority=49><w:lsdexception locked=false name="List Table 5 Dark Accent 6"priority=50><w:lsdexception locked=false name="List Table 6 Colorful Accent 6"priority=51><w:lsdexception locked=false name="List Table 7 Colorful Accent 6"priority=52><w:lsdexception locked=false name=Mention semihidden=true unhidewhenused=true><w:lsdexception locked=false name="Smart Hyperlink"semihidden=true unhidewhenused=true><w:lsdexception locked=false name=Hashtag semihidden=true unhidewhenused=true><w:lsdexception locked=false name="Unresolved Mention"semihidden=true unhidewhenused=true><w:lsdexception locked=false name="Smart Link"semihidden=true unhidewhenused=true></w:latentstyles></xml><![endif]--><style>@font-face{font-family:"Cambria Math";panose-1:2 4 5 3 5 4 6 3 2 4;mso-font-charset:0;mso-generic-font-family:roman;mso-font-pitch:variable;mso-font-signature:-536869121 1107305727 33554432 0 415 0}@font-face{font-family:Calibri;panose-1:2 15 5 2 2 2 4 3 2 4;mso-font-charset:0;mso-generic-font-family:swiss;mso-font-pitch:variable;mso-font-signature:-469750017 -1073732485 9 0 511 0}@font-face{font-family:Aptos;mso-font-charset:0;mso-generic-font-family:swiss;mso-font-pitch:variable;mso-font-signature:536871559 3 0 0 415 0}div.MsoNormal,li.MsoNormal,p.MsoNormal{mso-style-unhide:no;mso-style-qformat:yes;mso-style-parent:"";margin:0;mso-pagination:widow-orphan;font-size:12pt;font-family:"Times New Roman",serif;mso-fareast-font-family:"Times New Roman";mso-fareast-theme-font:minor-fareast}a:link,span.MsoHyperlink{mso-style-priority:99;font-family:"Times New Roman",serif;mso-bidi-font-family:"Times New Roman";color:#0563c1;mso-themecolor:hyperlink;text-decoration:underline;text-underline:single}a:visited,span.MsoHyperlinkFollowed{mso-style-noshow:yes;mso-style-priority:99;font-family:"Times New Roman",serif;mso-bidi-font-family:"Times New Roman";color:#954f72;mso-themecolor:followedhyperlink;text-decoration:underline;text-underline:single}div.msonormal0,li.msonormal0,p.msonormal0{mso-style-name:msonormal;mso-style-unhide:no;mso-margin-top-alt:auto;margin-right:0;mso-margin-bottom-alt:auto;margin-left:0;mso-pagination:widow-orphan;font-size:12pt;font-family:"Times New Roman",serif;mso-fareast-font-family:"Times New Roman";mso-fareast-theme-font:minor-fareast}span.SpellE{mso-style-name:"";mso-spl-e:yes}.MsoChpDefault{mso-style-type:export-only;mso-default-props:yes;font-size:10pt;mso-ansi-font-size:10pt;mso-bidi-font-size:10pt;mso-font-kerning:0;mso-ligatures:none}.MsoPapDefault{mso-style-type:export-only}@page WordSection1{size:595.3pt 841.9pt;margin:70.85pt 70.85pt 2cm 70.85pt;mso-header-margin:35.4pt;mso-footer-margin:35.4pt;mso-paper-source:0}div.WordSection1{page:WordSection1}</style><!--[if gte mso 10]><style>table.MsoNormalTable{mso-style-name:"Normale Tabelle";mso-tstyle-rowband-size:0;mso-tstyle-colband-size:0;mso-style-noshow:yes;mso-style-priority:99;mso-style-parent:"";mso-padding-alt:0 5.4pt 0 5.4pt;mso-para-margin:0;mso-pagination:widow-orphan;font-size:10pt;font-family:"Times New Roman",serif}</style><![endif]--><!--[if gte mso 9
		]><xml><o:shapedefaults spidmax=1026 v:ext=edit></xml><![endif]--><!--[if gte mso 9
		]><xml><o:shapelayout v:ext=edit><o:idmap data=1 v:ext=edit></o:shapelayout></xml><![endif]--><body lang=DE link=#0563C1 vlink=#954F72 style=tab-interval:35.4pt;word-wrap:break-word><div class=WordSection1><p class=MsoNormal><span style=font-size:11pt;mso-bidi-font-size:12pt;font-family:Calibri,sans-serif;mso-ascii-theme-font:minor-latin;mso-hansi-theme-font:minor-latin;mso-bidi-font-family:Aptos>Sehr geehrte Kundin, sehr geehrter Kunde,<o:p></o:p></span><p class=MsoNormal><span style=font-size:11pt;mso-bidi-font-size:12pt;font-family:Calibri,sans-serif;mso-ascii-theme-font:minor-latin;mso-hansi-theme-font:minor-latin;mso-bidi-font-family:Aptos><o:p> </o:p></span><p class=MsoNormal><span style=font-size:11pt;mso-bidi-font-size:12pt;font-family:Calibri,sans-serif;mso-ascii-theme-font:minor-latin;mso-hansi-theme-font:minor-latin;mso-bidi-font-family:Aptos>hiermit teilen wir Ihnen mit, dass Ihre Bestellung bei uns eingetroffen ist und aktuell unseren Wareneingang durchläuft.<o:p></o:p></span><p class=MsoNormal><span style=font-size:11pt;mso-bidi-font-size:12pt;font-family:Calibri,sans-serif;mso-ascii-theme-font:minor-latin;mso-hansi-theme-font:minor-latin;mso-bidi-font-family:Aptos><o:p> </o:p></span><p class=MsoNormal><span style=font-size:11pt;mso-bidi-font-size:12pt;font-family:Calibri,sans-serif;mso-ascii-theme-font:minor-latin;mso-hansi-theme-font:minor-latin;mso-bidi-font-family:Aptos>Sie können Ihre bestellte Ware ab dem nächsten Werktag 9 Uhr abholen.<o:p></o:p></span><p class=MsoNormal><span style=font-size:11pt;mso-bidi-font-size:12pt;font-family:Calibri,sans-serif;mso-ascii-theme-font:minor-latin;mso-hansi-theme-font:minor-latin;mso-bidi-font-family:Aptos><o:p> </o:p></span><p class=MsoNormal><span style=font-size:11pt;mso-bidi-font-size:12pt;font-family:Calibri,sans-serif;mso-ascii-theme-font:minor-latin;mso-hansi-theme-font:minor-latin;mso-bidi-font-family:Aptos>Unsere Öffnungszeiten:<o:p></o:p></span><p class=MsoNormal><span style=font-size:11pt;mso-bidi-font-size:12pt;font-family:Calibri,sans-serif;mso-ascii-theme-font:minor-latin;mso-hansi-theme-font:minor-latin;mso-bidi-font-family:Aptos>Montag - Freitag: 9:00 - 18:00<o:p></o:p></span><p class=MsoNormal><o:p> </o:p><table class=MsoNormalTable border=0 cellpadding=0 width=461 style="width:345.75pt;mso-cellspacing:1.5pt;mso-yfti-tbllook:1184;mso-padding-alt:0 5.4pt 0 5.4pt"><thead><tr style=mso-yfti-irow:0;mso-yfti-firstrow:yes><td style="padding:.75pt .75pt .75pt .75pt"><p class=MsoNormal><span style=font-size:11pt;font-family:Calibri,sans-serif;mso-ascii-theme-font:minor-latin;mso-hansi-theme-font:minor-latin;mso-bidi-font-family:Aptos;color:#000;mso-themecolor:text1>Mit freundlichen Grüßen<o:p></o:p></span><td style="padding:.75pt .75pt .75pt .75pt"></thead><tr style=mso-yfti-irow:1;mso-yfti-lastrow:yes><td style="border:none;border-right:solid #b7b7b7 1pt;mso-border-right-alt:solid #b7b7b7 .75pt;padding:.75pt .75pt .75pt .75pt"><table class=MsoNormalTable border=0 cellpadding=0 width=250 style="width:187.5pt;mso-cellspacing:1.5pt;mso-yfti-tbllook:1184;mso-padding-alt:0 5.4pt 0 5.4pt"><thead><tr style=mso-yfti-irow:0;mso-yfti-firstrow:yes><td style="padding:.75pt .75pt .75pt .75pt"><p class=MsoNormal><span style=font-size:16pt;font-family:Calibri,sans-serif;mso-ascii-theme-font:minor-latin;mso-hansi-theme-font:minor-latin;mso-bidi-font-family:Aptos;color:#0c509f>Ihr Computer Extra Team<o:p></o:p></span><td style="padding:.75pt .75pt .75pt .75pt"></thead><tr style=mso-yfti-irow:1;mso-yfti-lastrow:yes><td style="padding:.75pt .75pt .75pt .75pt"><p class=MsoNormal><span style=font-size:9pt;font-family:Calibri,sans-serif;mso-ascii-theme-font:minor-latin;mso-hansi-theme-font:minor-latin;mso-bidi-font-family:Aptos><o:p> </o:p></span><td style="padding:.75pt .75pt .75pt .75pt"></table><p class=MsoNormal><span style=font-family:Calibri,sans-serif;mso-ascii-theme-font:minor-latin;mso-hansi-theme-font:minor-latin;mso-bidi-font-family:Aptos;display:none;mso-hide:all><o:p> </o:p></span><table class=MsoNormalTable border=0 cellpadding=0 width=300 style="width:225pt;mso-cellspacing:1.5pt;mso-yfti-tbllook:1184;mso-padding-alt:0 5.4pt 0 5.4pt"><thead><tr style=mso-yfti-irow:0;mso-yfti-firstrow:yes><td style="padding:.75pt .75pt .75pt .75pt"><td style="padding:.75pt .75pt .75pt .75pt"></thead><tr style=mso-yfti-irow:1><td style="padding:.75pt .75pt .75pt .75pt"><p class=MsoNormal><span style=font-size:10pt;font-family:Calibri,sans-serif;mso-ascii-theme-font:minor-latin;mso-hansi-theme-font:minor-latin;mso-bidi-font-family:Aptos;color:#b7b7b7>Telefon<o:p></o:p></span><td style="padding:.75pt .75pt .75pt .75pt"><p class=MsoNormal><span style=font-size:10pt;font-family:Calibri,sans-serif;mso-ascii-theme-font:minor-latin;mso-hansi-theme-font:minor-latin;mso-bidi-font-family:Aptos;color:#0c509f><a href=tel:0561601440><span style=mso-bidi-font-family:Aptos;color:#0c509f>0561 60 144 0</span></a><o:p></o:p></span><tr style=mso-yfti-irow:2><td style="padding:.75pt .75pt .75pt .75pt"><p class=MsoNormal><span style=font-size:10pt;font-family:Calibri,sans-serif;mso-ascii-theme-font:minor-latin;mso-hansi-theme-font:minor-latin;mso-bidi-font-family:Aptos;color:#b7b7b7>Fax<o:p></o:p></span><td style="padding:.75pt .75pt .75pt .75pt"><p class=MsoNormal><span style=font-size:10pt;font-family:Calibri,sans-serif;mso-ascii-theme-font:minor-latin;mso-hansi-theme-font:minor-latin;mso-bidi-font-family:Aptos;color:#0c509f>0561 60 144 199<o:p></o:p></span><tr style=mso-yfti-irow:3><td style="padding:.75pt .75pt .75pt .75pt"><p class=MsoNormal><span style=font-size:10pt;font-family:Calibri,sans-serif;mso-ascii-theme-font:minor-latin;mso-hansi-theme-font:minor-latin;mso-bidi-font-family:Aptos;color:#b7b7b7>E-Mail<o:p></o:p></span><td style="padding:.75pt .75pt .75pt .75pt"><p class=MsoNormal><span style=font-size:10pt;font-family:Calibri,sans-serif;mso-ascii-theme-font:minor-latin;mso-hansi-theme-font:minor-latin;mso-bidi-font-family:Aptos;color:#0c509f><a href=mailto:service@computer-extra.de><span style=mso-bidi-font-family:Aptos;color:#0c509f>service@computer-extra.de</span></a><o:p></o:p></span><tr style=mso-yfti-irow:4><td style="padding:.75pt .75pt .75pt .75pt"><p class=MsoNormal><span style=font-size:10pt;font-family:Calibri,sans-serif;mso-ascii-theme-font:minor-latin;mso-hansi-theme-font:minor-latin;mso-bidi-font-family:Aptos;color:#b7b7b7>Webseite<o:p></o:p></span><td style="padding:.75pt .75pt .75pt .75pt"><p class=MsoNormal><span style=font-family:Calibri,sans-serif;mso-ascii-theme-font:minor-latin;mso-hansi-theme-font:minor-latin;mso-bidi-font-family:Aptos;color:#0c509f><a><span style=font-size:10pt;mso-bidi-font-family:Aptos;color:#0c509f>www.computer-extra.de</span></a></span><span style=font-size:10pt;font-family:Calibri,sans-serif;mso-ascii-theme-font:minor-latin;mso-hansi-theme-font:minor-latin;mso-bidi-font-family:Aptos;color:#0c509f><o:p></o:p></span><tr style=mso-yfti-irow:5;mso-yfti-lastrow:yes><td style="padding:.75pt .75pt .75pt .75pt"><td style="padding:.75pt .75pt .75pt .75pt"></table><p class=MsoNormal><span style=font-size:10pt;font-family:Calibri,sans-serif;mso-ascii-theme-font:minor-latin;mso-hansi-theme-font:minor-latin;mso-bidi-font-family:Aptos><o:p></o:p></span><td style="padding:.75pt .75pt .75pt .75pt"><table class=MsoNormalTable border=0 cellpadding=0 width=150 style="width:112.5pt;mso-cellspacing:1.5pt;mso-yfti-tbllook:1184;mso-padding-alt:0 5.4pt 0 5.4pt"><thead><tr style=mso-yfti-irow:0;mso-yfti-firstrow:yes><td style="padding:.75pt .75pt .75pt .75pt"><p class=MsoNormal><span style=font-size:14pt;font-family:Calibri,sans-serif;mso-ascii-theme-font:minor-latin;mso-hansi-theme-font:minor-latin;mso-bidi-font-family:Aptos;color:#2f5496;mso-themecolor:accent1;mso-themeshade:191;mso-no-proof:yes>Computer Extra GmbH</span><span style=font-size:10pt;font-family:Calibri,sans-serif;mso-ascii-theme-font:minor-latin;mso-hansi-theme-font:minor-latin;mso-bidi-font-family:Aptos;color:#2f5496;mso-themecolor:accent1;mso-themeshade:191><o:p></o:p></span></thead><tr style=mso-yfti-irow:1><td style="padding:.75pt .75pt .75pt .75pt"><tr style=mso-yfti-irow:2><td style="padding:.75pt .75pt .75pt .75pt"><tr style=mso-yfti-irow:3><td style="padding:.75pt .75pt .75pt .75pt"><p class=MsoNormal><span class=SpellE><span style=font-size:9pt;font-family:Calibri,sans-serif;mso-ascii-theme-font:minor-latin;mso-hansi-theme-font:minor-latin;mso-bidi-font-family:Aptos>Harleshäuser</span></span><span style=font-size:9pt;font-family:Calibri,sans-serif;mso-ascii-theme-font:minor-latin;mso-hansi-theme-font:minor-latin;mso-bidi-font-family:Aptos> Str. 8<o:p></o:p></span><tr style=mso-yfti-irow:4;mso-yfti-lastrow:yes><td style="padding:.75pt .75pt .75pt .75pt"><p class=MsoNormal><span style=font-size:9pt;font-family:Calibri,sans-serif;mso-ascii-theme-font:minor-latin;mso-hansi-theme-font:minor-latin;mso-bidi-font-family:Aptos>34130 Kassel<o:p></o:p></span><p class=MsoNormal><span style=font-size:9pt;font-family:Calibri,sans-serif;mso-ascii-theme-font:minor-latin;mso-hansi-theme-font:minor-latin;mso-bidi-font-family:Aptos><o:p> </o:p></span><p class=MsoNormal><span style=font-size:9pt;font-family:Calibri,sans-serif;mso-ascii-theme-font:minor-latin;mso-hansi-theme-font:minor-latin;mso-bidi-font-family:Aptos><o:p> </o:p></span><p class=MsoNormal><span style=font-size:9pt;font-family:Calibri,sans-serif;mso-ascii-theme-font:minor-latin;mso-hansi-theme-font:minor-latin;mso-bidi-font-family:Aptos><o:p> </o:p></span></table></table><p class=MsoNormal><span style=font-size:8pt;font-family:Calibri,sans-serif;mso-ascii-theme-font:minor-latin;mso-hansi-theme-font:minor-latin;mso-bidi-font-family:Aptos;color:#a5a5a5;mso-themecolor:accent3;mso-no-proof:yes>Sitz der Gesellschaft: 34130 Kassel<br>Geschäftsführer: Christian Krauss - Handelsregister: Kassel, HRB 19697<o:p></o:p></span><p class=MsoNormal><span style=font-size:8pt;font-family:Calibri,sans-serif;mso-ascii-theme-font:minor-latin;mso-hansi-theme-font:minor-latin;mso-bidi-font-family:Aptos;color:#a5a5a5;mso-themecolor:accent3;mso-no-proof:yes>USt.-IdNr.: DE357590630 - </span><span style=font-size:8pt;font-family:Calibri,sans-serif;mso-ascii-theme-font:minor-latin;mso-hansi-theme-font:minor-latin;mso-bidi-font-family:Aptos;color:#0c509f;mso-no-proof:yes><a><span style=mso-bidi-font-family:Aptos;color:#0c509f>Datenschutzinformationen</span></a></span><span style=font-size:8pt;font-family:Calibri,sans-serif;mso-ascii-theme-font:minor-latin;mso-hansi-theme-font:minor-latin;mso-bidi-font-family:Aptos;color:#a5a5a5;mso-themecolor:accent3;mso-no-proof:yes> - </span><span style=font-size:8pt;font-family:Calibri,sans-serif;mso-ascii-theme-font:minor-latin;mso-hansi-theme-font:minor-latin;mso-bidi-font-family:Aptos;color:#0c509f;mso-no-proof:yes><a><span style=mso-bidi-font-family:Aptos;color:#0c509f>AGB</span></a></span><span style=font-size:8pt;font-family:Calibri,sans-serif;mso-ascii-theme-font:minor-latin;mso-hansi-theme-font:minor-latin;mso-bidi-font-family:Aptos;color:#a5a5a5;mso-themecolor:accent3;mso-no-proof:yes> - </span><span style=font-size:8pt;font-family:Calibri,sans-serif;mso-ascii-theme-font:minor-latin;mso-hansi-theme-font:minor-latin;mso-bidi-font-family:Aptos;color:#0c509f;mso-no-proof:yes><a><span style=mso-bidi-font-family:Aptos;color:#0c509f>Impressum</span></a></span><span style=font-size:8pt;font-family:Calibri,sans-serif;mso-ascii-theme-font:minor-latin;mso-hansi-theme-font:minor-latin;mso-bidi-font-family:Aptos;color:#a5a5a5;mso-themecolor:accent3;mso-no-proof:yes><br style=mso-special-character:line-break><![if !supportLineBreakNewLine]><br style=mso-special-character:line-break><![endif]><o:p></o:p></span><p class=MsoNormal><span style=font-size:8pt;font-family:Calibri,sans-serif;mso-ascii-theme-font:minor-latin;mso-hansi-theme-font:minor-latin;mso-bidi-font-family:Aptos;color:#ff5d5d;mso-no-proof:yes>Der Inhalt dieser E-Mail und sämtliche Anhänge sind vertraulich und ausschließlich für den bezeichneten Empfänger bestimmt.<o:p></o:p></span><p class=MsoNormal><span style=font-size:8pt;font-family:Calibri,sans-serif;mso-ascii-theme-font:minor-latin;mso-hansi-theme-font:minor-latin;mso-bidi-font-family:Aptos;color:#ff5d5d;mso-no-proof:yes>Sollten Sie nicht der bezeichnete Empfänger sein, bitten wir Sie, umgehend den Absender zu benachrichtigen und diese E-Mail zu löschen.<o:p></o:p></span><p class=MsoNormal><span style=font-size:8pt;font-family:Calibri,sans-serif;mso-ascii-theme-font:minor-latin;mso-hansi-theme-font:minor-latin;mso-bidi-font-family:Aptos;color:#ff5d5d;mso-no-proof:yes>Jede Form der unautorisierten Veröffentlichung, Vervielfältigung und Weitergabe des Inhalts dieser E-Mail oder auch das Ergreifen von<o:p></o:p></span><p class=MsoNormal><span style=font-size:8pt;font-family:Calibri,sans-serif;mso-ascii-theme-font:minor-latin;mso-hansi-theme-font:minor-latin;mso-bidi-font-family:Aptos;color:#ff5d5d;mso-no-proof:yes>Maßnahmen als Reaktion darauf sind unzulässig.</span><span style=font-family:Calibri,sans-serif;mso-ascii-theme-font:minor-latin;mso-hansi-theme-font:minor-latin;mso-bidi-font-family:Aptos><o:p></o:p></span></div>
	`

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
	m.SetBody("text/html", body)

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

func syncAussteller() error {
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
	var query string = ""
	if len(Sage) > 0 {
		// Custom Query!
		query = "INSERT INTO Aussteller (id, Artikelnummer, Artikelname, Specs, Preis) VALUES"
		for i := range Sage {
			if i > 0 {
				query = fmt.Sprintf("%s,", query)
			}
			query = fmt.Sprintf("%s (%d, '%s', '%s', '%s', %.2f)", query, Sage[i].Id, Sage[i].Artikelnummer, strings.ReplaceAll(Sage[i].Artikelname, "'", "\""), strings.ReplaceAll(Sage[i].Specs, "'", "\""), Sage[i].Preis)
		}
		query = fmt.Sprintf("%s ON DUPLICATE KEY UPDATE Artikelname = VALUES(Artikelname), Specs = VALUES(Specs), Preis = VALUES(Preis);", query)
	}
	if query != "" {
		_, err := datebase.Exec(query)
		if err != nil {
			return err
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
