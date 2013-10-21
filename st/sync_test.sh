curl "localhost:7070/create?blockType=random"
curl "localhost:7070/create?blockType=sync"
curl "localhost:7070/blocks/2/set_rule" --data '{"Path":"t","Lag":20}'
curl "localhost:7070/connect?from=1&to=2"
curl "localhost:7070/create?blockType=tolog"
curl "localhost:7070/connect?from=2&to=4"
echo "things should start logging in ST in 20 seconds"
