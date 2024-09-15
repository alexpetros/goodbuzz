package db

import (
	"context"
	"log"

	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

type Tournament struct {
  tournament_id int64;
  name string;
}

func (t *Tournament) Id() int64 {
  return t.tournament_id
}

func (t *Tournament) Name() string {
  return t.name
}

type Room struct {
  room_id int64;
  name string;
}

func (r *Room) Id() int64 {
  return r.room_id
}

func (r *Room) Name() string {
  return r.name
}

var pool *sqlitex.Pool

func InitDb(filename string) {
	dbpool, err := sqlitex.NewPool("file:" + filename, sqlitex.PoolOptions{
		PoolSize: 10,
		PrepareConn: func(conn *sqlite.Conn) error {
			return sqlitex.ExecuteTransient(conn, "PRAGMA foreign_keys = ON;", nil)
		},
	})

	if err != nil {
		log.Fatal(err)
	}

	pool = dbpool
}

func Close() {
	pool.Close()
}

func GetTournament (ctx context.Context, id int64) *Tournament {
  conn, err := pool.Take(ctx)

  if err != nil {
    log.Printf("failed to take connection: %w\n", err)
  }

  defer pool.Put(conn)

  stmt := conn.Prep("SELECT tournament_id, name FROM tournaments WHERE tournament_id = $id")
  stmt.SetInt64("$id", id)

  row, err := stmt.Step()
  if err != nil {
    log.Printf("Error getting tournament: %s", err)
    return nil
  }
  if !row {
    log.Printf("No tournamnet found", err)
    return nil
  }

  tournament := Tournament {
    tournament_id: stmt.ColumnInt64(0),
    name: stmt.ColumnText(1),
  }

  stmt.Reset()
  return &tournament
}

func GetTournaments(ctx context.Context) []Tournament{
  conn, err := pool.Take(ctx)

  if err != nil {
    log.Printf("failed to take connection: %w\n", err)
  }

  defer pool.Put(conn)

  stmt := conn.Prep("SELECT tournament_id, name FROM tournaments")

  tournaments := make([]Tournament, 0)
  for {
    row, err := stmt.Step()
    if err != nil {
      log.Printf("Error getting tournaments: %s", err)
    }
    if !row {
      break
    }

    tournament := Tournament {
      tournament_id: stmt.ColumnInt64(0),
      name: stmt.ColumnText(1),
    }
    tournaments = append(tournaments, tournament)
  }

  return tournaments
}

func GetRoomsForTournament(ctx context.Context, tournament_id int64) []Room {
conn, err := pool.Take(ctx)

  if err != nil {
    log.Printf("failed to take connection: %w\n", err)
  }

  defer pool.Put(conn)

  stmt := conn.Prep("SELECT room_id, name FROM rooms WHERE tournament_id = $1")
  stmt.SetInt64("$1", tournament_id)

  rooms := make([]Room, 0)
  for {
    row, err := stmt.Step()
    if err != nil {
      log.Printf("Error getting tournaments: %s", err)
    }
    if !row {
      break
    }

    room := Room {
      room_id: stmt.ColumnInt64(0),
      name: stmt.ColumnText(1),
    }

    rooms = append(rooms, room)
  }

  return rooms
}
