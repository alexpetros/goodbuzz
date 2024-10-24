# Simple shell script for generating vegeta attacks
# I use the python one now but I'm leaving it here for reference
set -euo pipefail

URL="http://localhost:8080"

for room in {1..14}; do
  for player in {1..12}; do
    cat <<EOF
GET $URL/rooms/$room/player/live
Cookie: userToken=$(uuidgen)

EOF
  done
done
