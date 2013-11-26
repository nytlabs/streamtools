default=7070
PORT=${1:-$default}
echo -e "\033[0;36mcreating blocks\033[0m"
curl "localhost:${PORT}/create?blockType=ticker"
curl "localhost:${PORT}/create?blockType=random"
curl "localhost:${PORT}/create?blockType=tolog"
curl "localhost:${PORT}/create?blockType=count"
curl "localhost:${PORT}/"

echo -e "\033[0;36mtesting connections\033[0m"
#curl "localhost:${PORT}/connect?from=1&to=3"
curl "localhost:${PORT}/connect?from=2&to=4"
#curl "localhost:${PORT}/connect?from=2&to=4"
#sleep 2

#echo -e "\033[0;36mtesting routes\033[0m"
#curl "localhost:${PORT}/blocks/4/last_seen"
#curl "localhost:${PORT}/blocks/5/last_seen"
#curl "localhost:${PORT}/"

echo -e "\033[0;36mtesting count\033[0m"
curl --data '{"window":10}' "localhost:${PORT}/blocks/4/set_rule"
sleep 3
curl "localhost:${PORT}/blocks/4/count"

echo -e "\033[0;36mconnecting another block to count\033[0m"
curl "localhost:${PORT}/connect?from=1&to=4"
curl --data '{"window":15}' "localhost:${PORT}/blocks/4/set_rule"
sleep 3
curl "localhost:${PORT}/blocks/4/count"

echo -e "\033[0;36mbunching\033[0m"
curl "localhost:${PORT}/create?blockType=bunch"
curl "localhost:${PORT}/connect?from=2&to=7"
curl --data '{"Branch":"option", "EmitAfter":10}' "localhost:${PORT}/blocks/7/set_rule"
echo "connecting"
curl "localhost:${PORT}/connect?from=7&to=3"
