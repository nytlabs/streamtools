default=7070
PORT=${1:-$default}
clear
echo "tests filter block"
echo "all results in log"
echo "messages with status \"SHOULD PASS\" should be followed by a message."
echo "messages with status \"SHOULD FAIL\" should not followed by a message."
echo ""
echo ""
echo ""

curl "localhost:${PORT}/create?blockType=postto"
curl "localhost:${PORT}/create?blockType=filter"
curl "localhost:${PORT}/create?blockType=tolog"
curl "localhost:${PORT}/create?blockType=postto"

curl "localhost:${PORT}/blocks/2/set_rule" --data '{"Path":"test","Comparator":"test","Operator":"eq"}'
curl "localhost:${PORT}/connect?from=1&to=2"
curl "localhost:${PORT}/connect?from=2&to=3"
curl "localhost:${PORT}/connect?from=4&to=3"

echo ""
echo ""
echo "TEST 1: key:value"
echo ""
echo ""
read
curl "localhost:${PORT}/blocks/4/in" --data '{"STATUS":"SHOULD PASS"}'
curl "localhost:${PORT}/blocks/1/in" --data '{"test":"test"}'
curl "localhost:${PORT}/blocks/4/in" --data '{"STATUS":"SHOULD FAIL"}'
curl "localhost:${PORT}/blocks/1/in" --data '{"test":"ok"}'

curl "localhost:${PORT}/blocks/2/set_rule" --data '{"Path":"test","Comparator":5,"Operator":"eq"}'
curl "localhost:${PORT}/blocks/4/in" --data '{"STATUS":"SHOULD PASS"}'
curl "localhost:${PORT}/blocks/1/in" --data '{"test":5}'
curl "localhost:${PORT}/blocks/4/in" --data '{"STATUS":"SHOULD FAIL"}'
curl "localhost:${PORT}/blocks/1/in" --data '{"test":10}'


curl "localhost:${PORT}/blocks/2/set_rule" --data '{"Path":"test","Comparator":5,"Operator":"gt"}'
curl "localhost:${PORT}/blocks/4/in" --data '{"STATUS":"SHOULD PASS"}'
curl "localhost:${PORT}/blocks/1/in" --data '{"test":10}'
curl "localhost:${PORT}/blocks/4/in" --data '{"STATUS":"SHOULD FAIL"}'
curl "localhost:${PORT}/blocks/1/in" --data '{"test":5}'

curl "localhost:${PORT}/blocks/2/set_rule" --data '{"Path":"test","Comparator":true,"Operator":"eq"}'
curl "localhost:${PORT}/blocks/4/in" --data '{"STATUS":"SHOULD PASS"}'
curl "localhost:${PORT}/blocks/1/in" --data '{"test":true}'
curl "localhost:${PORT}/blocks/4/in" --data '{"STATUS":"SHOULD FAIL"}'
curl "localhost:${PORT}/blocks/1/in" --data '{"test":false}'

read
echo ""
echo ""
echo "TEST 2: key:array"
echo ""
echo ""
read
curl "localhost:${PORT}/blocks/2/set_rule" --data '{"Path":"test[]","Comparator":"one","Operator":"eq"}'
curl "localhost:${PORT}/blocks/4/in" --data '{"STATUS":"SHOULD PASS"}'
curl "localhost:${PORT}/blocks/1/in" --data '{"test":["one","two","three"]}'
curl "localhost:${PORT}/blocks/4/in" --data '{"STATUS":"SHOULD FAIL"}'
curl "localhost:${PORT}/blocks/1/in" --data '{"test":["two","three"]}'

curl "localhost:${PORT}/blocks/2/set_rule" --data '{"Path":"test[]","Comparator":5,"Operator":"gt"}'
curl "localhost:${PORT}/blocks/4/in" --data '{"STATUS":"SHOULD PASS"}'
curl "localhost:${PORT}/blocks/1/in" --data '{"test":[5,5,5,5,5,10]}'
curl "localhost:${PORT}/blocks/4/in" --data '{"STATUS":"SHOULD FAIL"}'
curl "localhost:${PORT}/blocks/1/in" --data '{"test":[1,2,3]}'

curl "localhost:${PORT}/blocks/2/set_rule" --data '{"Path":"test[]","Comparator":true,"Operator":"eq"}'
curl "localhost:${PORT}/blocks/4/in" --data '{"STATUS":"SHOULD PASS"}'
curl "localhost:${PORT}/blocks/1/in" --data '{"test":[true, false]}'
curl "localhost:${PORT}/blocks/4/in" --data '{"STATUS":"SHOULD FAIL"}'
curl "localhost:${PORT}/blocks/1/in" --data '{"test":[false, false, false]}'


read
echo ""
echo ""
echo "TEST 3: key:value subset"
echo ""
echo ""
read
curl "localhost:${PORT}/blocks/2/set_rule" --data '{"Path":"test","Comparator":"world","Operator":"subset"}'
curl "localhost:${PORT}/blocks/4/in" --data '{"STATUS":"SHOULD PASS"}'
curl "localhost:${PORT}/blocks/1/in" --data '{"test":"hello world"}'
curl "localhost:${PORT}/blocks/4/in" --data '{"STATUS":"SHOULD FAIL"}'
curl "localhost:${PORT}/blocks/1/in" --data '{"test":"hello"}'

read
echo ""
echo ""
echo "TEST 4: key:[{key:value}]"
echo ""
echo ""
read
curl "localhost:${PORT}/blocks/2/set_rule" --data '{"Path":"test[].sub","Comparator":"ok","Operator":"eq"}'
curl "localhost:${PORT}/blocks/4/in" --data '{"STATUS":"SHOULD PASS"}'
curl "localhost:${PORT}/blocks/1/in" --data '{"test":[{"sub":"ok"}]}'
curl "localhost:${PORT}/blocks/4/in" --data '{"STATUS":"SHOULD FAIL"}'
curl "localhost:${PORT}/blocks/1/in" --data '{"test":[{"sub": 5}]}'

curl "localhost:${PORT}/blocks/2/set_rule" --data '{"Path":"test[].sub","Comparator":5,"Operator":"eq"}'
curl "localhost:${PORT}/blocks/4/in" --data '{"STATUS":"SHOULD PASS"}'
curl "localhost:${PORT}/blocks/1/in" --data '{"test":[{"sub":5}]}'
curl "localhost:${PORT}/blocks/4/in" --data '{"STATUS":"SHOULD FAIL"}'
curl "localhost:${PORT}/blocks/1/in" --data '{"test":[{"sub":20}]}'

curl "localhost:${PORT}/blocks/2/set_rule" --data '{"Path":"test[].sub","Comparator":true,"Operator":"eq"}'
curl "localhost:${PORT}/blocks/4/in" --data '{"STATUS":"SHOULD PASS"}'
curl "localhost:${PORT}/blocks/1/in" --data '{"test":[{"sub":true}]}'
curl "localhost:${PORT}/blocks/4/in" --data '{"STATUS":"SHOULD FAIL"}'
curl "localhost:${PORT}/blocks/1/in" --data '{"test":[{"sub":false}]}'

read
echo ""
echo ""
echo "TEST 5: key:[{key:value}] subset"
echo ""
echo ""
read
curl "localhost:${PORT}/blocks/2/set_rule" --data '{"Path":"test[].sub","Comparator":"yell","Operator":"subset"}'
curl "localhost:${PORT}/blocks/4/in" --data '{"STATUS":"SHOULD PASS"}'
curl "localhost:${PORT}/blocks/1/in" --data '{"test":[{"sub":"yellow"}]}'
curl "localhost:${PORT}/blocks/4/in" --data '{"STATUS":"SHOULD FAIL"}'
curl "localhost:${PORT}/blocks/1/in" --data '{"test":[{"sub":"nope"}]}'

read
echo ""
echo ""
echo "TEST 6: array indices"
echo ""
echo ""
read
curl "localhost:${PORT}/blocks/2/set_rule" --data '{"Path":"test[0]","Comparator":0,"Operator":"eq"}'
curl "localhost:${PORT}/blocks/4/in" --data '{"STATUS":"SHOULD PASS"}'
curl "localhost:${PORT}/blocks/1/in" --data '{"test":[0,1,2,3]}'
curl "localhost:${PORT}/blocks/4/in" --data '{"STATUS":"SHOULD FAIL"}'
curl "localhost:${PORT}/blocks/1/in" --data '{"test":[null,"red",3]}'

curl "localhost:${PORT}/blocks/2/set_rule" --data '{"Path":"test[].sub[0]","Comparator":"exe","Operator":"subset"}'
curl "localhost:${PORT}/blocks/4/in" --data '{"STATUS":"SHOULD PASS"}'
curl "localhost:${PORT}/blocks/1/in" --data '{"test":[{"sub":["lexe","yes","no"]}]}'
curl "localhost:${PORT}/blocks/4/in" --data '{"STATUS":"SHOULD FAIL"}'
curl "localhost:${PORT}/blocks/1/in" --data '{"test":[{"sub":["no","lexe","no"]}]}'
curl "localhost:${PORT}/blocks/4/in" --data '{"STATUS":"SHOULD FAIL"}'
curl "localhost:${PORT}/blocks/1/in" --data '{"test":[]}'

curl "localhost:${PORT}/blocks/2/set_rule" --data '{"Path":"test[0].sub[]","Comparator":"exe","Operator":"subset"}'
curl "localhost:${PORT}/blocks/4/in" --data '{"STATUS":"SHOULD PASS"}'
curl "localhost:${PORT}/blocks/1/in" --data '{"test":[{"sub":["lexe","yes","no"]},{"sub":["ok","yes","no"]}]}'
curl "localhost:${PORT}/blocks/4/in" --data '{"STATUS":"SHOULD FAIL"}'
curl "localhost:${PORT}/blocks/1/in" --data '{"test":[{"sub":["no","yes","no"]},{"sub":["lexe","ok","no"]}]}'

read
echo ""
echo ""
echo "TEST 7: mixed"
echo ""
echo ""
read
curl "localhost:${PORT}/blocks/2/set_rule" --data '{"Path":"test[].sub[]","Comparator":"ok","Operator":"eq"}'
curl "localhost:${PORT}/blocks/4/in" --data '{"STATUS":"SHOULD PASS"}'
curl "localhost:${PORT}/blocks/1/in" --data '{"test":[{"sub":["ok","yes","no"]}]}'
curl "localhost:${PORT}/blocks/4/in" --data '{"STATUS":"SHOULD FAIL"}'
curl "localhost:${PORT}/blocks/1/in" --data '{"test":[{"sub":["no","yes","no"]}]}'
curl "localhost:${PORT}/blocks/4/in" --data '{"STATUS":"SHOULD FAIL"}'
curl "localhost:${PORT}/blocks/1/in" --data '{"test":[{"sub":null}]}'
curl "localhost:${PORT}/blocks/4/in" --data '{"STATUS":"SHOULD FAIL"}'
curl "localhost:${PORT}/blocks/1/in" --data '{"completely_wrong":null}'

curl "localhost:${PORT}/blocks/2/set_rule" --data '{"Path":"test[].sub[]","Comparator":"bees","Operator":"subset"}'
curl "localhost:${PORT}/blocks/4/in" --data '{"STATUS":"SHOULD PASS"}'
curl "localhost:${PORT}/blocks/1/in" --data '{"test":[{"sub":["bees are cool","yes","no"]}]}'
curl "localhost:${PORT}/blocks/4/in" --data '{"STATUS":"SHOULD FAIL"}'
curl "localhost:${PORT}/blocks/1/in" --data '{"test":[{"sub":["no","yes","no"]}]}'

curl "localhost:${PORT}/blocks/2/set_rule" --data '{"Path":"test[].sub[].url","Comparator":"fakeurl","Operator":"subset"}'
curl "localhost:${PORT}/blocks/4/in" --data '{"STATUS":"SHOULD PASS"}'
curl "localhost:${PORT}/blocks/1/in" --data '{"test":[{"sub":[{"url":"www.fakeurl.fake/~fake"},{"bad":null}]}]}'
curl "localhost:${PORT}/blocks/4/in" --data '{"STATUS":"SHOULD FAIL"}'
curl "localhost:${PORT}/blocks/1/in" --data '{"test":[{"sub":[null,5,"no"]}]}'


read
echo ""
echo ""
echo "TEST 8: null"
echo ""
echo ""
read
curl "localhost:${PORT}/blocks/2/set_rule" --data '{"Path":"test","Comparator":null,"Operator":"eq"}'
curl "localhost:${PORT}/blocks/4/in" --data '{"STATUS":"SHOULD PASS"}'
curl "localhost:${PORT}/blocks/1/in" --data '{"test":null}'
curl "localhost:${PORT}/blocks/4/in" --data '{"STATUS":"SHOULD FAIL"}'
curl "localhost:${PORT}/blocks/1/in" --data '{"test":5}'
curl "localhost:${PORT}/blocks/4/in" --data '{"STATUS":"SHOULD FAIL"}'
curl "localhost:${PORT}/blocks/1/in" --data '{"test":"no"}'
curl "localhost:${PORT}/blocks/4/in" --data '{"STATUS":"SHOULD FAIL"}'
curl "localhost:${PORT}/blocks/1/in" --data '{"test":false}'

curl "localhost:${PORT}/blocks/2/set_rule" --data '{"Path":"test[0]","Comparator":null,"Operator":"eq"}'
curl "localhost:${PORT}/blocks/4/in" --data '{"STATUS":"SHOULD PASS"}'
curl "localhost:${PORT}/blocks/1/in" --data '{"test":[null,1,2,3]}'
curl "localhost:${PORT}/blocks/4/in" --data '{"STATUS":"SHOULD FAIL"}'
curl "localhost:${PORT}/blocks/1/in" --data '{"test":["ok",null,2]}'

read
echo ""
echo ""
echo "TEST 9: escaped keys"
echo ""
echo ""
read
curl "localhost:${PORT}/blocks/2/set_rule" --data '{"Path":"[\"test.long.key\"]","Comparator":"test","Operator":"eq"}'
curl "localhost:${PORT}/blocks/4/in" --data '{"STATUS":"SHOULD PASS"}'
curl "localhost:${PORT}/blocks/1/in" --data '{"test.long.key":"test"}'
curl "localhost:${PORT}/blocks/4/in" --data '{"STATUS":"SHOULD FAIL"}'
curl "localhost:${PORT}/blocks/1/in" --data '{"test.long.key":"NOPE"}'

curl "localhost:${PORT}/blocks/2/set_rule" --data '{"Path":"test[\"foo.bar\"]","Comparator":5,"Operator":"eq"}'
curl "localhost:${PORT}/blocks/4/in" --data '{"STATUS":"SHOULD PASS"}'
curl "localhost:${PORT}/blocks/1/in" --data '{"test":{"foo.bar":5}}'
curl "localhost:${PORT}/blocks/4/in" --data '{"STATUS":"SHOULD FAIL"}'
curl "localhost:${PORT}/blocks/1/in" --data '{"test":10}'

curl "localhost:${PORT}/blocks/2/set_rule" --data '{"Path":"test[\"foo.bar\"][2]","Comparator":"test","Operator":"eq"}'
curl "localhost:${PORT}/blocks/4/in" --data '{"STATUS":"SHOULD PASS"}'
curl "localhost:${PORT}/blocks/1/in" --data '{"test":{"foo.bar":[0,1,"test"]}}'
curl "localhost:${PORT}/blocks/4/in" --data '{"STATUS":"SHOULD FAIL"}'
curl "localhost:${PORT}/blocks/1/in" --data '{"test":10}'

curl "localhost:${PORT}/blocks/2/set_rule" --data '{"Path":"[\"test.long.key\"][\"foo\"][\"bar\"].mixed","Comparator":"test","Operator":"eq"}'
curl "localhost:${PORT}/blocks/4/in" --data '{"STATUS":"SHOULD PASS"}'
curl "localhost:${PORT}/blocks/1/in" --data '{"test.long.key":{"foo":{"bar":{"mixed":"test"}}}}'
curl "localhost:${PORT}/blocks/2/set_rule" --data '{"Path":"[\"test.long.key\"].foo[\"bar\"].mixed","Comparator":"test","Operator":"eq"}'
curl "localhost:${PORT}/blocks/4/in" --data '{"STATUS":"SHOULD PASS"}'
curl "localhost:${PORT}/blocks/1/in" --data '{"test.long.key":{"foo":{"bar":{"mixed":"test"}}}}'
curl "localhost:${PORT}/blocks/2/set_rule" --data '{"Path":"[\"test.long.key\"][\"foo\"].bar[\"mixed\"]","Comparator":"test","Operator":"eq"}'
curl "localhost:${PORT}/blocks/4/in" --data '{"STATUS":"SHOULD PASS"}'
curl "localhost:${PORT}/blocks/1/in" --data '{"test.long.key":{"foo":{"bar":{"mixed":"test"}}}}'

echo ""
echo ""
echo "TEST 10: regex"
echo ""
echo ""
read

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
