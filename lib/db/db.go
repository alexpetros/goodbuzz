package db


import (
	"log"

	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

type Database struct {
	pool *sqlitex.Pool
}

var db Database

func InitDb(filename string) {
	pool, err := sqlitex.NewPool("file:" + filename, sqlitex.PoolOptions{
		PoolSize: 10,
		PrepareConn: func(conn *sqlite.Conn) error {
			return sqlitex.ExecuteTransient(conn, "PRAGMA foreign_keys = ON;", nil)
		},
	})

	if err != nil {
		log.Fatal(err)
	}

	db = Database{pool }
}

func Close() {
	db.pool.Close()
}

