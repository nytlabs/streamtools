default=7070
PORT=${1:-$default}
curl "localhost:${PORT}/create?blockType=random&id=r"
curl "localhost:${PORT}/create?blockType=groupHistogram&id=g"
curl "localhost:${PORT}/connect?from=r&to=g"
curl -s localhost:${PORT}/blocks/g/set_rule -d '{"Window":10, "GroupKey":"option", "Key":"c.nestedOption"}' 
sleep 4
curl -s localhost:${PORT}/blocks/g/list | jq .
curl -s localhost:${PORT}/blocks/g/histogram -d '{"GroupKey":"a"}' | jq .