default=7070
PORT=${1:-$default}
echo "creating random block"
curl "localhost:${PORT}/create?blockType=random&id=random"
echo ""
echo "creating mask"
curl "localhost:${PORT}/create?blockType=mask&id=mask"
echo ""
echo "setting rule for mask"
curl "localhost:${PORT}/blocks/mask/set_rule" --data '{"Mask":{"bad":{},"random_integers":{},"c":{"d":{}}}}'
echo ""
echo "creating tolog"
curl "localhost:${PORT}/create?blockType=tolog&id=log"
echo ""
echo "making connections"
echo ""
curl "localhost:${PORT}/connect?from=random&to=mask&id=random_to_mask"
curl "localhost:${PORT}/connect?from=mask&to=log&id=mask_to_log"
echo ""
echo "should be logging"
sleep 5
echo ""
echo "deleting connection"
curl "localhost:${PORT}/delete?id=random_to_mask"
echo ""
echo "should stop logging."
echo ""
echo "5"
sleep 1
echo "4"
sleep 1
echo "3"
sleep 1
echo "2"
sleep 1
echo "1"
sleep 1
curl "localhost:${PORT}/connect?from=random&to=mask&id=random_to_mask_2"
echo ""
echo "should be logging"
sleep 5
echo ""
echo "deleting a block"
echo ""
curl "localhost:${PORT}/delete?id=mask"
echo ""
echo "should stop logging"
echo ""
curl "localhost:${PORT}/create?blockType=mask&id=mask_2"
curl "localhost:${PORT}/blocks/mask_2/set_rule" --data '{"Mask":{"c":{"d":{}}}}'
echo ""
echo "setting a different rule for a new mask"
curl "localhost:${PORT}/connect?from=random&to=mask_2&id=random_to_mask_2"
curl "localhost:${PORT}/connect?from=mask_2&to=log&id=mask_to_log_2"
echo ""
echo "should be logging"
sleep 5
echo "deleting logger, should stop"
echo ""
curl "localhost:${PORT}/delete?id=log"
sleep 5
curl "localhost:${PORT}/create?blockType=tolog&id=log_2"
curl "localhost:${PORT}/connect?from=mask_2&to=log_2&id=mask_2_to_log_2"
echo ""
echo "should started logging again"
sleep 5
echo "deleting everything"
curl "localhost:${PORT}/delete?id=random"
curl "localhost:${PORT}/delete?id=log_2"
curl "localhost:${PORT}/delete?id=mask_2"
echo ""
echo ""
curl "localhost:${PORT}/"

