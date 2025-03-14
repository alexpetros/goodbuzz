package db

import (
	"context"
	"fmt"
	"goodbuzz/lib/logger"

	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

type Player struct {
	UserToken string
	Name      string
	Team      int64
	RoomId    int64
}

type Tournament struct {
	tournament_id int64
	name          string
	password      string
	num_rooms     int64
}

func (t *Tournament) Id() int64 {
	return t.tournament_id
}

func (t *Tournament) Name() string {
	return t.name
}

func (t *Tournament) Password() string {
	return t.password
}

func (t *Tournament) NumRooms() int64 {
	return t.num_rooms
}

func (t *Tournament) Url() string {
	return fmt.Sprintf("/tournaments/%d", t.tournament_id)
}

type Room struct {
	RoomId       int64
	TournamentId int64
	Name         string
	Description  string
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

	logger.Info("Connected to db at %s", filename)
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
		stmt := conn.Prep("SELECT tournament_id, name, password FROM tournaments WHERE tournament_id = $id")
		defer stmt.Reset()

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
			password:          stmt.ColumnText(2),
		}

		return &tournament
	}

	return run(ctx, fn)
}

func SetTournamentInfo(ctx context.Context, tournamentId int64, name string, password string) error {
	fn := func(conn *sqlite.Conn) error {
		stmt := conn.Prep("UPDATE tournaments SET name = $1, password = $2 WHERE tournament_id = $3")
		defer stmt.Reset()

		stmt.SetText("$1", name)
		stmt.SetText("$2", password)
		stmt.SetInt64("$3", tournamentId)

		_, err := stmt.Step()
		if err != nil {
			logger.Error("Failed to set tournament name: %s", err)
			return err
		}

		return nil
	}

	return run(ctx, fn)
}

func GetTournamentForRoom(ctx context.Context, roomId int64) *Tournament {
	fn := func(conn *sqlite.Conn) *Tournament {
		stmt := conn.Prep("SELECT tournament_id, tournaments.name FROM rooms LEFT JOIN tournaments USING (tournament_id) WHERE room_id = $id")
		defer stmt.Reset()

		stmt.SetInt64("$id", roomId)

		row, err := stmt.Step()
		if err != nil {
			logger.Error("Error getting tournament: %w", err)
			return nil
		}
		if !row {
			logger.Warn("Room %d not found found", roomId)
			return nil
		}

		tournament := Tournament{
			tournament_id: stmt.ColumnInt64(0),
			name:          stmt.ColumnText(1),
		}

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
		defer stmt.Reset()

		tournaments := make([]Tournament, 0)
		for {
			row, err := stmt.Step()
			if err != nil {
				logger.Error("Error getting tournaments: %w", err)
			}
			if !row {
				break
			}

			tournament := Tournament{
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

func CreateTournament(ctx context.Context, name string) error {
	fn := func(conn *sqlite.Conn) error {
		stmt := conn.Prep("INSERT INTO tournaments (name) VALUES ($1)")
		defer stmt.Reset()

		stmt.SetText("$1", name)

		_, err := stmt.Step()
		if err != nil {
			logger.Error("%v", err)
			return err
		}

		return nil
	}

	return run(ctx, fn)
}

func CreateRoom(ctx context.Context, tournament_id int64, name string) error {
	fn := func(conn *sqlite.Conn) error {
		stmt := conn.Prep("INSERT INTO rooms (name, tournament_id) VALUES ($1, $2)")
		defer stmt.Reset()

		stmt.SetText("$1", name)
		stmt.SetInt64("$2", tournament_id)

		_, err := stmt.Step()
		if err != nil {
			logger.Error("%v", err)
			return err
		}

		return nil
	}

	return run(ctx, fn)
}

func GetRoom(ctx context.Context, room_id int64) *Room {
	fn := func(conn *sqlite.Conn) *Room {
		stmt := conn.Prep("SELECT room_id, name, description, tournament_id FROM rooms WHERE room_id = $1")
		defer stmt.Reset()

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
			RoomId:       stmt.ColumnInt64(0),
			Name:         stmt.ColumnText(1),
			Description:  stmt.ColumnText(2),
			TournamentId: stmt.ColumnInt64(3),
		}

		return &room
	}
	return run(ctx, fn)
}

func SetRoomNameAndDescription(ctx context.Context, roomId int64, name string, description string) error {
	fn := func(conn *sqlite.Conn) error {
		stmt := conn.Prep("UPDATE rooms SET description = $1, name = $2 WHERE room_id = $3")
		defer stmt.Reset()

		stmt.SetText("$1", description)
		stmt.SetText("$2", name)
		stmt.SetInt64("$3", roomId)

		_, err := stmt.Step()
		if err != nil {
			logger.Error("Failed to set room description: %s", err)
			return err
		}

		return nil
	}

	return run(ctx, fn)
}

func DeleteRoom(ctx context.Context, roomId int64) error {
	fn := func(conn *sqlite.Conn) error {
		stmt := conn.Prep("DELETE FROM rooms WHERE room_id = $1")
		defer stmt.Reset()

		stmt.SetInt64("$1", roomId)

		_, err := stmt.Step()
		if err != nil {
			logger.Error("Failed to delete room %d: %s", roomId, err)
			return err
		}

		return nil
	}

	return run(ctx, fn)
}

func GetRoomsForTournament(ctx context.Context, tournament_id int64) []Room {
	fn := func(conn *sqlite.Conn) []Room {
		stmt := conn.Prep("SELECT room_id, name, description FROM rooms WHERE tournament_id = $1")
		defer stmt.Reset()

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
				RoomId:      stmt.ColumnInt64(0),
				Name:        stmt.ColumnText(1),
				Description: stmt.ColumnText(2),
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
		defer stmt.Reset()

		stmt.SetInt64("$1", tournament_id)

		_, err := stmt.Step()
		if err != nil {
			logger.Error("Failed to delete tournament: %s", err)
			return err
		}

		return nil
	}

	return run(ctx, fn)
}

func CreatePlayer(ctx context.Context, userToken string, name string, team int64, roomId int64) error {
	fn := func(conn *sqlite.Conn) error {
		stmt := conn.Prep("INSERT OR REPLACE INTO players (user_token, name, team, room_id) VALUES ($1, $2, $3, $4)")
		defer stmt.Reset()

		stmt.SetText("$1", userToken)
		stmt.SetText("$2", name)
		stmt.SetInt64("$3", team)
		stmt.SetInt64("$4", roomId)

		_, err := stmt.Step()
		if err != nil {
			logger.Error("Failed to create player: %s", err)
			return err
		}

		return nil
	}

	return run(ctx, fn)
}

func GetPlayer(ctx context.Context, userToken string) *Player {
	fn := func(conn *sqlite.Conn) *Player {
		stmt := conn.Prep("SELECT user_token, name, team, room_id FROM players WHERE user_token = $1")
		defer stmt.Reset()

		stmt.SetText("$1", userToken)

		row, err := stmt.Step()
		if err != nil {
			logger.Error("Error getting user: %s", err)
			return nil
		}
		if !row {
			return nil
		}

		return &Player{
			UserToken: stmt.ColumnText(0),
			Name:      stmt.ColumnText(1),
			Team:      stmt.ColumnInt64(2),
			RoomId:    stmt.ColumnInt64(3),
		}
	}

	return run(ctx, fn)
}

func UpdatePlayer(ctx context.Context, userToken string, name string, team int64) error {
	fn := func(conn *sqlite.Conn) error {
		stmt := conn.Prep("UPDATE players SET name = $1, team = $2 WHERE user_token = $3")
		defer stmt.Reset()

		stmt.SetText("$1", name)
		stmt.SetInt64("$2", team)
		stmt.SetText("$3", userToken)

		_, err := stmt.Step()
		if err != nil {
			logger.Error("Failed to update player: %s", err)
			return err
		}

		return nil
	}

	return run(ctx, fn)
}

func SetUserName(ctx context.Context, userToken string, name string) error {
	fn := func(conn *sqlite.Conn) error {
		stmt := conn.Prep("UPDATE players SET name = $1 WHERE user_token = $2")
		defer stmt.Reset()

		stmt.SetText("$1", name)
		stmt.SetText("$2", userToken)

		_, err := stmt.Step()
		if err != nil {
			logger.Error("Failed to set name: %s", err)
			return err
		}

		return nil
	}

	return run(ctx, fn)
}

func LoginMod(ctx context.Context, userToken string) error {
	fn := func(conn *sqlite.Conn) error {
		stmt := conn.Prep("INSERT OR REPLACE INTO mod_sessions (user_token) VALUES ($1)")
		defer stmt.Reset()

		stmt.SetText("$1", userToken)

		_, err := stmt.Step()
		if err != nil {
			logger.Error("%v", err)
			return err
		}

		return nil
	}

	return run(ctx, fn)
}

func LoginAdmin(ctx context.Context, userToken string) error {
	fn := func(conn *sqlite.Conn) error {
		stmt := conn.Prep("INSERT OR REPLACE INTO admin_sessions (user_token) VALUES ($1)")
		defer stmt.Reset()

		stmt.SetText("$1", userToken)

		_, err := stmt.Step()
		if err != nil {
			logger.Error("%v", err)
			return err
		}

		return nil
	}

	return run(ctx, fn)
}

func DeleteLogin(ctx context.Context, userToken string) error {
	fn := func(conn *sqlite.Conn) error {
		modDelete := conn.Prep("DELETE FROM mod_sessions WHERE user_token = $1")
		defer modDelete.Reset()
		adminDelete := conn.Prep("DELETE FROM admin_sessions WHERE user_token = $1")
		defer adminDelete.Reset()

		modDelete.SetText("$1", userToken)
		adminDelete.SetText("$1", userToken)

		modDelete.Step()
		adminDelete.Step()

		return nil
	}

	return run(ctx, fn)
}

func IsMod(ctx context.Context, userToken string) bool {
	fn := func(conn *sqlite.Conn) bool {
		stmt := conn.Prep("SELECT 1 FROM mod_sessions WHERE user_token = $1")
		defer stmt.Reset()

		stmt.SetText("$1", userToken)

		row, err := stmt.Step()
		if err != nil {
			logger.Error("Error attempting to get moderator status %s", err)
			return false
		}
		if !row {
			return false
		}

		res := stmt.ColumnInt(0)

		return res == 1
	}

	return run(ctx, fn)
}

func IsAdmin(ctx context.Context, userToken string) bool {
	fn := func(conn *sqlite.Conn) bool {
		stmt := conn.Prep("SELECT 1 FROM admin_sessions WHERE user_token = $1")
		defer stmt.Reset()

		stmt.SetText("$1", userToken)

		row, err := stmt.Step()
		if err != nil {
			logger.Error("Error attempting to get moderator status %s", err)
			return false
		}
		if !row {
			return false
		}

		res := stmt.ColumnInt(0)

		return res == 1
	}

	return run(ctx, fn)
}

func ModPassword(ctx context.Context) string {
	fn := func(conn *sqlite.Conn) string {
		stmt := conn.Prep("SELECT value FROM settings WHERE key = 'mod_password'")
		defer stmt.Reset()

		row, err := stmt.Step()
		if err != nil {
			panic("Panicked whil looking for admin password")
		}
		if !row {
			panic("Missing admin password - this is a setup issue")
		}

		return stmt.ColumnText(0)
	}

	return run(ctx, fn)
}

func AdminPassword(ctx context.Context) string {
	fn := func(conn *sqlite.Conn) string {
		stmt := conn.Prep("SELECT value FROM settings WHERE key = 'admin_password'")
		defer stmt.Reset()

		row, err := stmt.Step()
		if err != nil {
			panic("Panicked whil looking for admin password")
		}
		if !row {
			panic("Missing admin password - this is a setup issue")
		}

		return stmt.ColumnText(0)
	}

	return run(ctx, fn)
}

func SetSetting(ctx context.Context, key string, value string) error {
	fn := func(conn *sqlite.Conn) error {
		stmt := conn.Prep("UPDATE settings SET value = $1 WHERE key = $2")
		defer stmt.Reset()

		stmt.SetText("$1", value)
		stmt.SetText("$2", key)

		_, err := stmt.Step()
		if err != nil {
			logger.Error("Failed to set key: %s\n%v", err, err)
			return err
		}

		return nil
	}

	return run(ctx, fn)
}

func WipeSessions(ctx context.Context, userToken string) error {
	fn := func(conn *sqlite.Conn) error {
		modDelete := conn.Prep("DELETE FROM mod_sessions")
		defer modDelete.Reset()
		// Delete all the sessions except that of the user doing the deleting
		adminDelete := conn.Prep("DELETE FROM admin_sessions WHERE user_token != $1")
		defer adminDelete.Reset()

		adminDelete.SetText("$1", userToken)

		modDelete.Step()
		adminDelete.Step()

		return nil
	}

	return run(ctx, fn)
}

func AddUserToTournament(ctx context.Context, userToken string, tournament_id int64) error {
	fn := func(conn *sqlite.Conn) error {
		stmt := conn.Prep("INSERT INTO player_logins (user_token, tournament_id) VALUES ($1, $2)")
		defer stmt.Reset()

		stmt.SetText("$1", userToken)
		stmt.SetInt64("$2", tournament_id)
		stmt.Step()

		return nil
	}

	return run(ctx, fn)
}


func IsUserAuthedForTournament(ctx context.Context, userToken string, tournament_id int64) bool {
	fn := func(conn *sqlite.Conn) bool {
		stmt := conn.Prep("SELECT 1 FROM player_logins WHERE user_token = $1")
		defer stmt.Reset()

		stmt.SetText("$1", userToken)

		row, err := stmt.Step()
		if err != nil {
			logger.Error("Error attempting to get player tournament info %s", err)
			return false
		}
		if !row {
			return false
		}

		res := stmt.ColumnInt(0)

		return res == 1
	}

	return run(ctx, fn)
}
