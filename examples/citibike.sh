curl "localhost:7070/create?blockType=getHTTP&id=citibike"
curl localhost:7070/blocks/citibike/set_rule -d '{"Endpoint":"http://citibikenyc.com/stations/json"}'

curl "localhost:7070/create?blockType=ticker&id=tock"
curl localhost:7070/blocks/tock/set_rule -d '{"Interval":"4s"}'

curl "localhost:7070/connect?from=tock&to=citibike"

curl "localhost:7070/create?blockType=map&id=availableBikes"
curl "localhost:7070/blocks/availableBikes/set_rule" -d '{"Map":{"availableBikes":".stationBeanList[].availableBikes"}}'

curl "localhost:7070/connect?from=citibike&to=availableBikes"

curl "localhost:7070/create?blockType=tolog&id=logger"
curl "localhost:7070/connect?from=availableBikes&to=logger"

