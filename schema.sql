CREATE TABLE tournaments (
  tournament_id INTEGER PRIMARY KEY,
  name TEXT NOT NULL,
  url TEXT NOT NULL
) STRICT;

CREATE TABLE rooms (
  room_id INTEGER PRIMARY KEY,
  name TEXT NOT NULL,
  tournament_id INTEGER REFERENCES tournaments ON UPDATE CASCADE ON DELETE CASCADE
) STRICT;

-- Some basic test data
INSERT INTO tournaments (name, url)
VALUES
  ('October Online Tournament', 'october-online'),
  ('November Online Tournament', 'november-online');

-- Some basic test data
INSERT INTO rooms (name, tournament_id)
VALUES
  ('Happy Hippo', 1),
  ('Grumpy Gorilla', 1),
  ('Daring Dog', 2);

