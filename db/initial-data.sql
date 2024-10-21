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


