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

