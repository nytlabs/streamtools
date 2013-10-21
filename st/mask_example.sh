curl "localhost:7070/create?blockType=mask"
curl "localhost:7070/blocks/1/set_rule" --data '{"Mask":{"a":{},"c":{"d":{}}}}'
curl "localhost:7070/create?blockType=random"
curl "localhost:7070/connect?from=2&to=1"
curl "localhost:7070/create?blockType=tolog"
curl "localhost:7070/connect?from=1&to=4"
