echo -e "\033[0;36mcreating blocks\033[0m"
curl "localhost:7070/create?blockType=ticker"
curl "localhost:7070/create?blockType=random"
curl "localhost:7070/create?blockType=tolog"
curl "localhost:7070/create?blockType=count"
curl "localhost:7070/"

echo -e "\033[0;36mtesting connections\033[0m"
#curl "localhost:7070/connect?from=1&to=3"
curl "localhost:7070/connect?from=2&to=4"
#curl "localhost:7070/connect?from=2&to=4"
#sleep 2

#echo -e "\033[0;36mtesting routes\033[0m"
#curl "localhost:7070/blocks/4/last_seen"
#curl "localhost:7070/blocks/5/last_seen"
#curl "localhost:7070/"

echo -e "\033[0;36mtesting count\033[0m"
curl --data '{"window":10}' "localhost:7070/blocks/4/set_rule"
sleep 3
curl "localhost:7070/blocks/4/count"

echo -e "\033[0;36mconnecting another block to count\033[0m"
curl "localhost:7070/connect?from=1&to=4"
curl --data '{"window":15}' "localhost:7070/blocks/4/set_rule"
sleep 3
curl "localhost:7070/blocks/4/count"

echo -e "\033[0;36mbunching\033[0m"
curl "localhost:7070/create?blockType=bunch"
curl "localhost:7070/connect?from=2&to=7"
curl --data '{"Branch":"option", "EmitAfter":10}' "localhost:7070/blocks/7/set_rule"
echo "connecting"
curl "localhost:7070/connect?from=7&to=3"
