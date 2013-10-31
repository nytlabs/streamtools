curl 'localhost:7070/create?blockType=random'
curl 'localhost:7070/create?blockType=postValue'
curl 'localhost:7070/create?blockType=tolog'

curl -d '{"Endpoint":"http://localhost:7070/blocks/1/set_rule", "keyMapping":[{"MsgKey": "random_int", "QueryKey": "Period"}] }' 'localhost:7070/blocks/2/set_rule'

curl 'localhost:7070/connect?from=1&to=2'
curl 'localhost:7070/connect?from=1&to=3'
