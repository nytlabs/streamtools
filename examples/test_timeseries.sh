curl "localhost:7070/create?blockType=random&id=r"
curl "localhost:7070/create?blockType=timeseries&id=t"
curl localhost:7070/blocks/t/set_rule --data '{"NumSamples":10, "Key":"random_float"}'
curl "localhost:7070/connect?from=r&to=t"

curl localhost:7070/blocks/t/set_rule --data '{"NumSamples":10, "Key":"random_float", "Lag": 1800}'
