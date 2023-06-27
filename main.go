package main

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"time"

	"github.com/asaskevich/govalidator"
	_ "modernc.org/sqlite"
)

// uniqid returns a unique id string useful when generating random strings.
// It was lifted from https://www.php2golang.com/method/function.uniqid.html.
func uniqid(prefix string) string {
	now := time.Now()
	sec := now.Unix()
	usec := now.UnixNano() % 0x100000

	return fmt.Sprintf("%s%08x%05x", prefix, sec, usec)
}

type Shortener interface {
	Shorten() string
}

type URLShortener struct {
	long, short string
}

func (s *URLShortener) Shorten() string {
	var (
		randomChars   = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0987654321")
		randIntLength = 27
		stringLength  = 32
	)

	str := make([]rune, stringLength)

	for char := range str {
		nBig, err := rand.Int(rand.Reader, big.NewInt(int64(randIntLength)))
		if err != nil {
			panic(err)
		}

		str[char] = randomChars[nBig.Int64()]
	}

	hash := sha256.Sum256([]byte(uniqid(string(str))))
	encodedString := base64.StdEncoding.EncodeToString(hash[:])

	return encodedString[0:9]
}

type App struct {
	db        *sql.DB
	shortener Shortener
}

func newApp() App {
	db, err := sql.Open("sqlite", "data/database.sqlite3")
	if err != nil {
		log.Fatal(err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	return App{db: db}
}

func (app *App) shortenUrl(writer http.ResponseWriter, request *http.Request) {
	request.ParseForm()

	longUrl := request.FormValue("url")
	if longUrl == "" {
		http.Error(writer, "URL was not provided or not able to be retrieved from the request.", http.StatusBadRequest)
		return
	}

	if !govalidator.IsURL(longUrl) {
		http.Error(writer, fmt.Sprintf("URL [%s] was not valid.", longUrl), http.StatusBadRequest)
		return
	}

	parsedURL, err := url.Parse(longUrl)
	if err != nil {
		http.Error(writer, fmt.Sprintf("URL [%s] was not valid: %v", longUrl, err), http.StatusBadRequest)
		return
	}
	shortUrl := fmt.Sprintf("%s://%s", parsedURL.Scheme, app.shortener.Shorten())

	result, err := app.db.Exec("INSERT INTO urls(short, long) VALUES(?, ?)", shortUrl, longUrl)
	if err != nil {
		http.Error(writer, fmt.Sprintf("could not insert shortened URL: %v", err), 500)
		return
	}
	_, err = result.RowsAffected()
	if err != nil {
		http.Error(writer, fmt.Sprintf("could not insert shortened URL: %v", err), 500)
		return
	}

	writer.Write([]byte(fmt.Sprintf("%s was shortened to %s", longUrl, shortUrl)))
}

func hasQueryParameterMiddleware(next http.Handler, parameter string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get(parameter) == "" {
			http.Error(w, fmt.Sprintf("Query parameter [%s] not available.", parameter), 400)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (app *App) getURL(writer http.ResponseWriter, request *http.Request) {
	request.ParseForm()

	url := request.FormValue("url")
	if url == "" {
		http.Error(writer, "URL was not provided or not able to be retrieved from the request.", http.StatusBadRequest)
		return
	}

	fmt.Println("looking for a long URL for", url)
	row := app.db.QueryRow("SELECT short, long FROM urls WHERE short = $1", url)
	shortener := new(URLShortener)
	err := row.Scan(&shortener.short, &shortener.long)
	if err == sql.ErrNoRows {
		fmt.Println("could not find a long URL for", url)
		http.NotFound(writer, request)
		return
	}
	if err != nil {
		http.Error(
			writer,
			fmt.Sprintf("something went wrong looking up a long URL for %s: %v\n", url, err),
			http.StatusInternalServerError,
		)
		return
	}

	fmt.Printf("found a long URL (%s) for %s\n", shortener.long, shortener.long)
	http.Redirect(writer, request, shortener.long, http.StatusMovedPermanently)
}

func main() {
	app := newApp()

	http.Handle("/", hasQueryParameterMiddleware(http.HandlerFunc(app.shortenUrl), "url",))
	http.Handle("/get", hasQueryParameterMiddleware(http.HandlerFunc(app.getURL), "url"))

	http.ListenAndServe(":8080", nil)
}
