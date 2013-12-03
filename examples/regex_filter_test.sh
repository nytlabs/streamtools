default=7070
PORT=${1:-$default}
clear

curl "localhost:${PORT}/create?blockType=postto"
curl "localhost:${PORT}/create?blockType=filter"
curl "localhost:${PORT}/create?blockType=tolog"
curl "localhost:${PORT}/create?blockType=postto"

curl "localhost:${PORT}/blocks/2/set_rule" --data
'{"Path":"test","Comparator":"test","Operator":"eq"}'
curl "localhost:${PORT}/connect?from=1&to=2"
curl "localhost:${PORT}/connect?from=2&to=3"
curl "localhost:${PORT}/connect?from=4&to=3"

curl "localhost:${PORT}/blocks/2/set_rule" --data '{"Path":"test","Comparator":"bob","Operator":"regex"}'
curl "localhost:${PORT}/blocks/4/in" --data '{"STATUS":"SHOULD PASS"}'
curl "localhost:${PORT}/blocks/1/in" --data '{"test":"bob"}'
curl "localhost:${PORT}/blocks/4/in" --data '{"STATUS":"SHOULD FAIL"}'
curl "localhost:${PORT}/blocks/1/in" --data '{"test":"bo"}'


curl "localhost:${PORT}/blocks/2/set_rule" --data '{"Path":"test","Comparator":"http://www.nytimes.com/$","Operator":"regex"}'
curl "localhost:${PORT}/blocks/4/in" --data '{"STATUS":"SHOULD PASS"}'
curl "localhost:${PORT}/blocks/1/in" --data '{"test":"http://www.nytimes.com/"}'
curl "localhost:${PORT}/blocks/4/in" --data '{"STATUS":"SHOULD FAIL"}'
curl "localhost:${PORT}/blocks/1/in" --data '{"test":"http://www.nytimes.com/bob"}'
