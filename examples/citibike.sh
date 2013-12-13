curl "localhost:7070/create?blockType=get&id=citibike"
curl localhost:7070/blocks/citibike/set_rule -d '{"Endpoint":"http://citibikenyc.com/stations/json"}'
curl "localhost:7070/create?blockType=ticker&id=t"
curl localhost:7070/blocks/t/set_rule -d '{"Period":4}'
curl "localhost:7070/connect?from=t&to=citibike"
curl "localhost:7070/create?blockType=unpack&id=unpacker"
curl localhost:7070/blocks/unpacker/set_rule -d '{"Path":"stationBeanList[]"}'
curl "localhost:7070/connect?from=citibike&to=unpacker"
curl "localhost:7070/create?blockType=tolog&id=l"
#curl "localhost:7070/connect?from=unpacker&to=l"
