package lib

import (
	"log"
	"os"

	"zombiezen.com/go/sqlite/sqlitex"
)

func GetDb(filename string) *sqlitex.Pool {
  dbpool, err := sqlitex.NewPool("file:" + filename, sqlitex.PoolOptions{
    PoolSize: 10,
  })

  if err != nil {
    log.Fatal(err)
    os.Exit(1)
  }

  return dbpool
}
