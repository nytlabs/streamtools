default=7070
PORT=${1:-$default}
curl "localhost:${PORT}/create?blockType=random&id=r"
curl "localhost:${PORT}/create?blockType=filter&id=f"
curl "localhost:${PORT}/create?blockType=mask&id=m"

curl "localhost:${PORT}/create?blockType=tolog&id=l"

curl "localhost:${PORT}/blocks/f/set_rule" --data '{"Path":"[\"sometimes.dot.option\"]","Operator":"keyin","Comparator":"","Invert":false}'

curl "localhost:${PORT}/blocks/m/set_rule" --data '{"Mask":{"sometimes.dot.option":{},"t":{}}}'

curl "localhost:${PORT}/connect?from=r&to=f"
curl "localhost:${PORT}/connect?from=f&to=m"
curl "localhost:${PORT}/connect?from=m&to=l"
