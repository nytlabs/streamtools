default=7070
PORT=${1:-$default}
curl "localhost:${PORT}/create?blockType=random&id=rnd"
curl "localhost:${PORT}/create?blockType=mask&id=mask"
curl "localhost:${PORT}/blocks/mask/set_rule" --data '{"Mask":{"t":{},"random_integers":{}}}'
curl "localhost:${PORT}/connect?from=rnd&to=mask&id=rnd_to_mask"
curl "localhost:${PORT}/create?blockType=sync&id=sync"
curl "localhost:${PORT}/blocks/sync/set_rule" --data '{"Path":"t","Lag":20}'
curl "localhost:${PORT}/connect?from=mask&to=sync&id=mask_to_sync"
curl "localhost:${PORT}/create?blockType=tolog&id=log"
curl "localhost:${PORT}/connect?from=sync&to=log&id=sync_to_log"
echo ""
echo ""
echo "items should start logging in ~20 seconds."
echo "press key to end"
read
echo ""
echo ""
curl "localhost:${PORT}/delete?id=rnd"
curl "localhost:${PORT}/delete?id=mask"
curl "localhost:${PORT}/delete?id=sync"
curl "localhost:${PORT}/delete?id=log"
echo ""
echo ""
curl "localhost:${PORT}"
echo ""
echo ""
echo "done."
