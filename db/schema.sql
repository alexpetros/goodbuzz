CREATE TABLE tournaments (
  tournament_id INTEGER PRIMARY KEY,
  name TEXT NOT NULL,
  password TEXT NOT NULL DEFAULT "",
) STRICT;

CREATE TABLE rooms (
  room_id INTEGER PRIMARY KEY,
  name TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT "",
  tournament_id INTEGER REFERENCES tournaments ON UPDATE CASCADE ON DELETE CASCADE
) STRICT;

CREATE TABLE admin_sessions (
  user_token TEXT PRIMARY KEY,
  timestamp TEXT DEFAULT CURRENT_TIMESTAMP
) STRICT;

CREATE TABLE mod_sessions (
  user_token TEXT PRIMARY KEY,
  timestamp TEXT DEFAULT CURRENT_TIMESTAMP
) STRICT;

CREATE TABLE player_sessions (
  user_token TEXT NOT NULL,
  tournament_id INTEGER NOT NULL,
  timestamp TEXT DEFAULT CURRENT_TIMESTAMP
) STRICT;

CREATE TABLE players (
  user_token TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  team INTEGER NOT NULL,
  room_id INTEGER NOT NULL REFERENCES rooms ON UPDATE CASCADE ON DELETE CASCADE,
  timestamp TEXT DEFAULT CURRENT_TIMESTAMP
) STRICT;

CREATE TABLE settings (
  key TEXT PRIMARY KEY,
  value TEXT NOT NULL
) STRICT;


INSERT INTO SETTINGS (key, value) VALUES
('admin_password', 'admin'),
('mod_password', 'moderator');
