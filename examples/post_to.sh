default=7070
PORT=${1:-$default}
curl "localhost:${PORT}/create?blockType=postto"
curl "localhost:${PORT}/create?blockType=tolog"
curl "localhost:${PORT}/connect?from=1&to=2"
curl --data '{"DATA":"TEST"}' "localhost:${PORT}/blocks/1/in"
curl --data '{"DATA":"TEST2"}' "localhost:${PORT}/blocks/1/in"
