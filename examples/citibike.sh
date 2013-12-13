curl "localhost:7070/create?blockType=get&id=citibike"
curl "localhost:7070/create?blockType=tolog&id=logger"
curl "localhost:7070/connect?from=citibike&to=logger"
curl "localhost:7070/blocks/citibike/set_rule" -d '{"Endpoint":"http://citibikenyc.com/stations/json"}'
curl "localhost:7070/create?blockType=ticker&id=ticker"
curl "localhost:7070/blocks/ticker/set_rule" -d '{"Period":4}'
curl "localhost:7070/connect?from=ticker&to=citibike"
