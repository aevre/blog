package database

import (
	"net/url"
	"testing"

	"blog-server/internal/platform/config"
)

func TestDSNEncoding(t *testing.T) {
	for _, password := range []string{"", "a b'@:/?#"} {
		cfg := &config.Config{Database: config.DatabaseConfig{Host: "localhost", Port: 5432, User: "reader", Password: password, Name: "disposable"}}
		u, err := url.Parse(NewDSN(cfg))
		if err != nil {
			t.Fatal(err)
		}
		got, _ := u.User.Password()
		if got != password || u.Path != "/disposable" {
			t.Fatalf("incorrect DSN fields")
		}
	}
}
