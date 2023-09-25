package history

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/mattn/go-sqlite3"
	"github.com/solodov/org-roam-alfred-items/alfred"
)

var Path string

type Record struct {
	Ts      int64
	Trigger string
	Item    string
}

const schema = `CREATE TABLE items (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  ts INTEGER NOT NULL,
  trigger VARCHAR(32) NOT NULL,
  query VARCHAR(128) NOT NULL,
  item VARCHAR(1024) NOT NULL);
PRAGMA auto_vacuum = "incremental";
PRAGMA incremental_vacuum(10);`

func Open() (*sql.DB, error) {
	// TODO: create directory tree
	initDb := false
	if _, err := os.Stat(Path); err != nil {
		log.Print("history database doesn't exist, initializing...")
		initDb = true
	}
	db, err := sql.Open("sqlite3_extended", Path)
	if err != nil {
		return nil, err
	}
	if initDb {
		if _, err = db.Exec(schema); err != nil {
			return nil, err
		}
	}
	return db, nil
}

func init() {
	regex := func(re, s string) (bool, error) {
		return regexp.MatchString(re, s)
	}
	sql.Register("sqlite3_extended",
		&sqlite3.SQLiteDriver{
			ConnectHook: func(conn *sqlite3.SQLiteConn) error {
				return conn.RegisterFunc("regexp", regex, true)
			},
		})
}

func FindMatchingItems(trigger, alfredQuery string) (items []alfred.Item) {
	db, err := Open()
	if err != nil {
		log.Println("failed to open history database: %v", err)
		return items
	}
	regex := strings.Join(strings.Split(alfredQuery, " "), "|")
	dbQuery := `
    SELECT ts, item FROM items
    WHERE trigger = ? AND query REGEXP ?
    ORDER BY ts DESC LIMIT 40`
	row, err := db.Query(dbQuery, trigger, regex)
	if err != nil {
		log.Printf("history database query failed: %v\n", err)
		return items
	}
	seen := map[string]bool{}
	for row.Next() {
		var ts int64
		var itemStr string
		if err = row.Scan(&ts, &itemStr); err != nil {
			log.Printf("history db row scan failed: %v\n", err)
			continue
		}
		if _, found := seen[itemStr]; found {
			continue
		}
		seen[itemStr] = true
		var item alfred.Item
		if err := json.Unmarshal([]byte(itemStr), &item); err != nil {
			log.Printf("invalid item json: %v\n", err)
			continue
		}
		item.Title = fmt.Sprintf("%s: %s", formatDuration(time.Now().Sub(time.Unix(ts, 0))), item.Title)
		items = append(items, item)
	}
	return items
}

const (
	hour  = 60 * time.Minute
	day   = 24 * hour
	week  = 7 * day
	month = 31 * week
)

func formatDuration(d time.Duration) string {
	if d > month {
		return "month+"
	}
	if d > week {
		return "week+"
	}
	if d > day {
		return "day+"
	}
	if d > hour {
		return "hour+"
	}
	if d > time.Minute {
		return "minute+"
	}
	return "-minute"
}
