CREATE TABLE tournaments (
  tournament_id INTEGER PRIMARY KEY,
  name TEXT NOT NULL,
  url TEXT NOT NULL
) STRICT;

CREATE TABLE rooms (
  room_id INTEGER PRIMARY KEY,
  name TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT "",
  tournament_id INTEGER REFERENCES tournaments ON UPDATE CASCADE ON DELETE CASCADE
) STRICT;

CREATE TABLE admin_sessions (
  user_token TEXT PRIMARY KEY,
  timestamp INTEGER DEFAULT CURRENT_TIMESTAMP
) STRICT;

CREATE TABLE mod_sessions (
  user_token TEXT PRIMARY KEY,
  timestamp INTEGER DEFAULT CURRENT_TIMESTAMP
) STRICT;

CREATE TABLE game_users (
  user_token TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  timestamp INTEGER DEFAULT CURRENT_TIMESTAMP
) STRICT;

