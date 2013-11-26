default=7070
PORT=${1:-$default}
curl "localhost:${PORT}/"
echo $PORT

