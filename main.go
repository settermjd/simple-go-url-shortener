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

// ShortenURL generates and returns a short URL string.
func ShortenURL() string {
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

type URLShortener struct {
	long, short string
}

type App struct {
	db *sql.DB
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

	parsedURL, err := url.Parse(longUrl)
	if err != nil {
		http.Error(writer, fmt.Sprintf("URL was not valid: %s", err), http.StatusBadRequest)
		return
	}
	shortUrl := fmt.Sprintf("%s://%s", parsedURL.Scheme, ShortenURL())

	result, err := app.db.Exec("INSERT INTO urls(short, long) VALUES($1, $2)", shortUrl, longUrl)
	if err != nil {
		http.Error(writer, http.StatusText(500), 500)
		return
	}
	_, err = result.RowsAffected()
	if err != nil {
		http.Error(writer, http.StatusText(500), 500)
		return
	}

	writer.Write([]byte(fmt.Sprintf("%s was shortened to %s", longUrl, shortUrl)))
}

func main() {

	app := newApp()

	http.HandleFunc("/", app.shortenUrl)
	http.ListenAndServe(":8080", nil)
}
