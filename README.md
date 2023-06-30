# Simple Go URL Shortener

This is a small example application that shows how to build a rather rudimentary URL shortener in Go. 
It also forms the basis for an upcoming [Twilio tutorial](https://www.twilio.com/blog/author/msetter) that I'm writing.
It's not meant to be a project that you would base anything on.

## Getting Started

To get started, after cloning the codebase, create a new file in the _data_ directory, named _database.sqlite3_.
Then, through whatever database tool you prefer to use, such as [the SQLite command line shell](https://www.sqlite.org/cli.html), run the following query:

```sql
PRAGMA foreign_keys=OFF;
BEGIN TRANSACTION;
CREATE TABLE urls (
    short TEXT NOT NULL,
    long TEXT NOT NULL
);
CREATE UNIQUE INDEX idx_urls ON urls(short, long);
COMMIT;
```

## Usage

Start the application running with the following command:

```bash
go run main.go
```

To shorten a URL, run the following command, replacing the url-encoded URL with one of your choice.

```bash
curl -i http://localhost:8080/\?url\=http%3A%2F%2Flocalhost%3A3000
```

To retrieve the long/original URL for a shortened URL, run the following command, replacing the url-encoded URL with one that you shortened.

```bash
curl --location -i http://localhost:8080/get\?url\=https%3A%2F%2FWKQw2COtp
```

## Questions

If you have questions or queries, get in touch: matthew[at]matthewsetter.com.
