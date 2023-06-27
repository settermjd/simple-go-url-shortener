package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

type SpyURLShortener struct {
}

func (s *SpyURLShortener) Shorten() string {
	return "short"
}

func TestShortenUrl(t *testing.T) {
	t.Run("can handle the url being invalid", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, "/?url=blah", nil)
		response := httptest.NewRecorder()

		app := App{}
		app.shortenUrl(response, request)
		got := response.Body.String()
		expected := "URL [blah] was not valid.\n"

		if got != expected {
			t.Errorf("received: %q, expected: %q", got, expected)
		}
	})

	t.Run("can handle the url not being present", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, "/", nil)
		response := httptest.NewRecorder()

		app := App{}
		app.shortenUrl(response, request)
		got := response.Body.String()
		expected := "URL was not provided or not able to be retrieved from the request.\n"

		if got != expected {
			t.Errorf("received: %q, expected: %q", got, expected)
		}
	})

	t.Run("can shorten and store shortened url if a valid url is present", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, "/?url=http%3A%2F%2Flocalhost%3A8080", nil)
		response := httptest.NewRecorder()

		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer db.Close()

		var longURL, shortURL = "http://localhost:8080", "http://short"
		mock.ExpectExec("INSERT INTO urls").
			WithArgs(shortURL, longURL).
			WillReturnResult(sqlmock.NewResult(1, 1))

		shortener := &SpyURLShortener{}
		app := App{db: db, shortener: shortener}
		app.shortenUrl(response, request)
		got := response.Body.String()
		expected := fmt.Sprintf("%s was shortened to %s", longURL, shortURL)

		if got != expected {
			t.Errorf("received: %q, expected: %q", got, expected)
		}
	})

	t.Run("can handle the required parameter not being present", func(t *testing.T) {
		param := "url"
		handlerToTest := hasQueryParameterMiddleware(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}), 
			param,
		)
		req := httptest.NewRequest("GET", "/", nil)
		res := httptest.NewRecorder()
		handlerToTest.ServeHTTP(res, req)
		got := res.Body.String()
		expected := fmt.Sprintf("Query parameter [%s] not available.\n", param)

		if got != expected {
			t.Errorf("received: %q, expected: %q", got, expected)
		}
	})
}
