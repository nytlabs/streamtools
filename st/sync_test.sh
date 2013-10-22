curl "localhost:7070/create?blockType=random&id=s_random"
curl "localhost:7070/create?blockType=sync&id=s_sync"
curl "localhost:7070/blocks/s_sync/set_rule" --data '{"Path":"t","Lag":20}'
curl "localhost:7070/connect?from=s_random&to=s_sync"
curl "localhost:7070/create?blockType=tolog&id=s_log"
curl "localhost:7070/connect?from=s_sync&to=s_log"
echo "things should start logging in ST in 20 seconds"
