package db

import (
	"context"
	"fmt"
	"log"

	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

type Tournament struct {
	tournament_id int64
	name          string
}

func (t *Tournament) Id() int64 {
	return t.tournament_id
}

func (t *Tournament) Name() string {
	return t.name
}

func (t *Tournament) Url() string {
	return fmt.Sprintf("/tournaments/%d", t.tournament_id)
}

func (t *Tournament) EditUrl() string {
	return fmt.Sprintf("/tournaments/%d/edit", t.tournament_id)
}

type Room struct {
	room_id int64
	name    string
}

func (r *Room) Id() int64 {
	return r.room_id
}

func (r *Room) Name() string {
	return r.name
}

var pool *sqlitex.Pool

func InitDb(filename string) {
	dbpool, err := sqlitex.NewPool("file:"+filename, sqlitex.PoolOptions{
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

type dbFn[T any] func(conn *sqlite.Conn) T

func run[T any](ctx context.Context, fn dbFn[T]) T {
	conn, err := pool.Take(ctx)

	if err != nil {
		log.Printf("failed to take connection: %w\n", err)
	}

	defer pool.Put(conn)

	return fn(conn)
}

func GetTournament(ctx context.Context, id int64) *Tournament {
	fn := func(conn *sqlite.Conn) *Tournament {
		stmt := conn.Prep("SELECT tournament_id, name FROM tournaments WHERE tournament_id = $id")
		stmt.SetInt64("$id", id)

		row, err := stmt.Step()
		if err != nil {
			log.Printf("Error getting tournament: %s", err)
			return nil
		}
		if !row {
			log.Printf("No tournament found", err)
			return nil
		}

		tournament := Tournament{
			tournament_id: stmt.ColumnInt64(0),
			name:          stmt.ColumnText(1),
		}

		stmt.Reset()
		return &tournament
	}

	return run(ctx, fn)
}

func GetTournaments(ctx context.Context) []Tournament {
	fn := func(conn *sqlite.Conn) []Tournament {
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

			tournament := Tournament{
				tournament_id: stmt.ColumnInt64(0),
				name:          stmt.ColumnText(1),
			}
			tournaments = append(tournaments, tournament)
		}

		return tournaments
	}

	return run(ctx, fn)
}

func GetRoom(ctx context.Context, room_id int64) *Room {
	fn := func(conn *sqlite.Conn) *Room {
		stmt := conn.Prep("SELECT room_id, name FROM rooms WHERE room_id = $1")
		stmt.SetInt64("$1", room_id)

		row, err := stmt.Step()
		if err != nil {
			log.Printf("Error getting tournament: %s", err)
			return nil
		}
		if !row {
			log.Printf("No tournamnet found", err)
			return nil
		}

		room := Room{
			room_id: stmt.ColumnInt64(0),
			name:    stmt.ColumnText(1),
		}

		stmt.Reset()
		return &room
	}
	return run(ctx, fn)
}

func GetRoomsForTournament(ctx context.Context, tournament_id int64) []Room {
	fn := func(conn *sqlite.Conn) []Room {
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

			room := Room{
				room_id: stmt.ColumnInt64(0),
				name:    stmt.ColumnText(1),
			}

			rooms = append(rooms, room)
		}

		return rooms
	}
	return run(ctx, fn)
}

func DeleteTournament(ctx context.Context, tournament_id int64) error {
	fn := func(conn *sqlite.Conn) error {
		stmt := conn.Prep("DELETE FROM tournaments WHERE tournament_id = $1")
		stmt.SetInt64("$1", tournament_id)

		_, err := stmt.Step()
		if err != nil {
			log.Printf("Error deleting tournament: %s", err)
			return err
		}

		stmt.Reset()
		return nil
	}

	return run(ctx, fn)
}
