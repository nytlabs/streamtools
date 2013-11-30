curl localhost:7070/create?blockType=get
curl localhost:7070/create?blockType=tolog
curl "localhost:7070/connect?from=1&to=2"
curl localhost:7070/blocks/1/set_rule -d '{"Endpoint":"http://citibikenyc.com/stations/json"}'
curl localhost:7070/create?blockType=ticker
curl localhost:7070/blocks/4/set_rule -d '{"Period":4}'
curl "localhost:7070/connect?from=4&to=1"
