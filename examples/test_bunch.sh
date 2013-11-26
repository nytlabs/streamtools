curl "localhost:7070/create?blockType=random"
curl "localhost:7070/create?blockType=bunch"
curl "localhost:7070/blocks/2/set_rule" -d '{"Branch":"option", "EmitAfter":10}'
curl "localhost:7070/create?blockType=tofile"
curl "localhost:7070/blocks/3/set_rule" -d '{"Filename":"testBunch.json"}'
curl "localhost:7070/connect?from=1&to=2"
curl "localhost:7070/connect?from=2&to=3"

sleep 12
tail -f ~/testBunch.json | jq .bunch[].option
