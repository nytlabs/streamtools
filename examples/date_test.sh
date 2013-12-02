default=7070
PORT=${1:-$default}
curl "localhost:${PORT}/create?blockType=date"
curl "localhost:${PORT}/create?blockType=tolog"
curl "localhost:${PORT}/blocks/1/set_rule" --data '{"FmtString":"3:04PM","Period":1}'
curl "localhost:${PORT}/connect?from=1&to=2"
