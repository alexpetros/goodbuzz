# Goodbuzz

An online, simultaneous, high-capacity buzzer system, primarily intended for quiz bowl competitions (like Jeopardy).

You enter to a room, someone reads a question (on a video call, elsewhere), and when you think you know the answer, you click the "Buzz" button.
The room's moderator can either reset the buzzer for everyone, or everyone but the person who buzzed in.
The app shows you everyone who's in the room with you, and doesn't require a login.

It's not yet hosted publicly, but it will be once I make a couple modifications to support that.

## Installation

This uses [templ](https://github.com/a-h/templ) for templating and [air](https://github.com/air-verse/air) for live-reloading development.
Follow the installation instructions for both.

Once these are installed, use the makefile to build and run the program:

- `make` / `make live` - start the hot-reloading dev version
- `make build` - build the production version
- `make dev` - build and run the dev  version
- `make prod` - build and run the production  version
- `make clean` - remove all build artifacts

I found the hot-reloading somewhat problematic, and often found myself just using `make dev`.

