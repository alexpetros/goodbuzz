set -euo pipefail

for i in {1..100}; do
  cat <<EOF
GET http://localhost:8080/rooms/1/player/live
Cookie: userToken=$(uuidgen)

EOF
done
