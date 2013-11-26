default=7070
PORT=${1:-$default}
echo "localhost:${PORT}/create?blockType=random"
curl "localhost:${PORT}/create?blockType=random"
curl "localhost:${PORT}/create?blockType=postValue"
curl "localhost:${PORT}/create?blockType=tolog"
curl -d "{\"Endpoint\":\"http://localhost:${PORT}/blocks/1/set_rule\", \"keyMapping\":[{\"MsgKey\": \"random_int\", \"QueryKey\": \"Period\"}] }" "localhost:${PORT}/blocks/2/set_rule"
curl "localhost:${PORT}/connect?from=1&to=2"
curl "localhost:${PORT}/connect?from=1&to=3"