package db

import (
	"context"
	"fmt"
	"goodbuzz/lib/logger"

	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

type Tournament struct {
	tournament_id int64
	name          string
	num_rooms     int64
}

func (t *Tournament) Id() int64 {
	return t.tournament_id
}

func (t *Tournament) Name() string {
	return t.name
}

func (t *Tournament) NumRooms() int64 {
	return t.num_rooms
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
		logger.Fatal("Failed to initialize database: %w", err)
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
		logger.Error("Failed to take connection: %w", err)
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
			logger.Error("Error getting tournament: %w", err)
			return nil
		}
		if !row {
			logger.Warn("Tournament %d not found found", id)
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
		stmt := conn.Prep(`
			SELECT tournament_id, tournaments.name, count(room_id) as num_rooms
			FROM tournaments
			LEFT JOIN rooms USING (tournament_id)
			GROUP BY tournament_id
		`)

		tournaments := make([]Tournament, 0)
		for {
			row, err := stmt.Step()
			if err != nil {
				logger.Error("Error getting tournaments: %w", err)
			}
			if !row {
				break
			}

			tournament := Tournament {
				tournament_id: stmt.ColumnInt64(0),
				name:          stmt.ColumnText(1),
				num_rooms:     stmt.ColumnInt64(2),
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
			logger.Error("Error getting room: %s", err)
			return nil
		}
		if !row {
			logger.Warn("Room %d not found", room_id)
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
				logger.Error("Error getting rooms for tournament: %s", err)
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
			logger.Error("Failed to delete tournament: %s", err)
			return err
		}

		stmt.Reset()
		return nil
	}

	return run(ctx, fn)
}
