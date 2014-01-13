default=7070
PORT=${1:-$default}
curl "localhost:${PORT}/create?blockType=random&id=r"
curl "localhost:${PORT}/create?blockType=toRedis&id=redis"
curl "localhost:${PORT}/create?blockType=tolog&id=log"

curl "localhost:${PORT}/create?blockType=count&id=c"
curl "localhost:${PORT}/blocks/c/set_rule" --data '{"Window":"5s"}' 

# connect random output to the counter
curl "localhost:${PORT}/connect?from=r&to=c"

curl "localhost:${PORT}/create?blockType=ticker&id=t"
curl "localhost:${PORT}/blocks/t/set_rule" --data '{"Interval":"10s"}'

# connect the ticket to the counter's poll endpoint
curl "localhost:${PORT}/connect?from=t&to=c/poll"

# store output of counter in redis and print to stdout
curl "localhost:${PORT}/connect?from=c&to=redis"
curl "localhost:${PORT}/connect?from=r&to=log"
curl "localhost:${PORT}/connect?from=c&to=log"
