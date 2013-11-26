default=7070
PORT=${1:-$default}
echo "creating mask block"
curl "localhost:${PORT}/create?blockType=mask&id=m_mask"
echo "setting rule for mask"
curl "localhost:${PORT}/blocks/m_mask/set_rule" --data '{"Mask":{"a":{},"c":{"d":{}}}}'
echo "creating random bock"
curl "localhost:${PORT}/create?blockType=random&id=m_random"
echo "connecting random to mask"
curl "localhost:${PORT}/connect?from=m_random&to=m_mask"
echo "creating tolog block"
curl "localhost:${PORT}/create?blockType=tolog&id=m_log"
echo "connecting mask to log"
curl "localhost:${PORT}/connect?from=m_mask&to=m_log"
