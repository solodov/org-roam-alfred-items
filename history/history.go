package history

import (
	"database/sql"
	"encoding/json"
	"log"
	"os"
	"regexp"
	"strings"

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
	dbQuery := `SELECT ts, item FROM items WHERE trigger = ? AND query REGEXP ? ORDER BY ts LIMIT 40`
	row, err := db.Query(dbQuery, trigger, regex)
	if err != nil {
		log.Println("history database query failed: %v", err)
		return items
	}
	for row.Next() {
		var ts int64
		var itemStr string
		err := row.Scan(&ts, &itemStr)
		if err != nil {
			log.Println("history db row scan failed: %v", err)
			continue
		}
		var item alfred.Item
		err = json.Unmarshal([]byte(itemStr), &item)
		if err != nil {
			log.Println("invalid item json: %v", err)
			continue
		}
		// when := time.Now().Sub(time.Unix(ts, 0))
		item.Title = "history: " + item.Title
		items = append(items, item)
	}
	return items
}
