# Goodbuzz

An online, simultaneous, high-capacity buzzer system, primarily intended for quiz bowl competitions (like Jeopardy).

You enter to a room, someone reads a question (on a video call, elsewhere), and when you think you know the answer, you click the "Buzz" button.
The room's moderator can either reset the buzzer for everyone, or everyone but the person who buzzed in.
The app shows you everyone who's in the room with you, and doesn't require a login.

It's not yet hosted publicly, but it will be once I make a couple modifications to support that.

## Developer Setup

This uses [templ](https://github.com/a-h/templ) for templating, so make sure that the `templ` command is installed and accessible from the shell

Use the makefile to build and run the program:

- `make` / `make build` - build the production version
- `make dev` - build and run the dev  version
- `make prod` - build and run the production  version
- `make clean` - remove all build artifacts

I found the hot-reloading somewhat problematic, and often found myself just using `make dev`.

## Running with Docker

The easiest way to run the app locally is with Docker Compose.

### Prerequisites

- [Docker](https://docs.docker.com/get-docker/)
- [Docker Compose](https://docs.docker.com/compose/install/) (included with Docker Desktop)

### Start the app

From the project root:

```bash
docker compose up --build
```

This will:

- Build the app image from the `Dockerfile`
- Start the container on port **8080**
- Create a `data/` folder on your machine and mount it into the container
- Initialize the SQLite database automatically on first run

Open the app at [http://localhost:8080](http://localhost:8080).

### Run in the background

```bash
docker compose up --build -d
```

### View logs

```bash
docker compose logs -f
```

### Stop the app

```bash
docker compose down
```

The database is stored in `./data/goodbuzz.db` on your host, so tournament and room data persists across restarts.

### Build and run without Compose

```bash
docker build -t goodbuzz .
docker run --rm -p 8080:8080 -v "$(pwd)/data:/app/data" goodbuzz
```

## Server Installation

Included with the source is a shell script, `./ubuntu-vps-setup.sh`, that sets up the program on a Virtual Private Server (VPS) running Ubuntu.
I use DigitalOcean, but as long as you're on the LTS Ubuntu version indicated in the script, you should be fine.

To install it on a fresh VPS, run `cat ./ubuntu-vps-setup.sh | ssh roo@YOUR_VPS_IP`.
The script will great a `goweb` user that will run the application, and copy the `root` public keys so that anyone who can ssh into `root` can ssh to `goweb`.
