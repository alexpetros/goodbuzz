#!python3
# Stress-test the server
# Requires "vegeta" installation
import sys
import uuid
import subprocess

# URL = "http://localhost:8080"
URL = "https://goodbuzz.iacompetitions.com"
NUM_PLAYERS = 13

def makeModerator(room):
    token = uuid.uuid4()
    req = f"GET {URL}/rooms/{room}/moderator/live\n"
    req += f"Cookie: userToken={token}\n"
    return req

def makePlayer(room):
    token = uuid.uuid4()
    req = f"GET {URL}/rooms/{room}/player/live\n"
    req += f"Cookie: userToken={token}\n"
    return req

startRoom = 1
endRoom = 15

if len(sys.argv) > 1:
    startRoom = int(sys.argv[1])

if len(sys.argv) > 2:
    endRoom = int(sys.argv[2])

# Make the attack list and send it
connections = []
for room in range(startRoom, endRoom):
    for player in range(1, NUM_PLAYERS + 1):
        connections.append(makePlayer(room))
    connections.append(makeModerator(room))

attacks = "\n".join(connections)
maxCon = len(connections)

subprocess.run(
        ["vegeta", "attack", "-duration=2s", "-timeout=100s", f"-max-connections={maxCon}"],
        input=attacks,
        encoding="ascii"
        )
