package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"time"
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

type App struct {

}

func newApp() App {
	return App{}
}

func (app *App) shortenUrl(writer http.ResponseWriter, request *http.Request) {
	request.ParseForm()

	longUrl := request.FormValue("url")
	if longUrl == "" {
		writer.WriteHeader(400)
		writer.Write([]byte("URL was not provided or not able to be retrieved from the request."))
		return
	}

	// validate the URL
	parsedURL, err := url.Parse(longUrl)
	if err != nil {
		writer.WriteHeader(400)
		writer.Write([]byte("URL was not valid."))
		return
	}
	shortUrl := ShortenURL()
	writer.Write([]byte(fmt.Sprintf(
		"%s was shortened to %s://%s", 
		longUrl, 
		parsedURL.Scheme, 
		shortUrl,
	)))
}

func main() {
	app := newApp()

	http.HandleFunc("/", app.shortenUrl)
	http.ListenAndServe(":8080", nil)
}