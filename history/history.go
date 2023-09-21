package history

import (
	"database/sql"
	"log"
	"os"
	"regexp"

	"github.com/mattn/go-sqlite3"
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
