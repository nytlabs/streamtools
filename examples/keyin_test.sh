curl "localhost:8080/create?blockType=random&id=r"
curl "localhost:8080/create?blockType=filter&id=f"
curl "localhost:8080/create?blockType=mask&id=m"

curl "localhost:8080/create?blockType=tolog&id=l"

curl "localhost:8080/blocks/f/set_rule" --data '{"Path":"[\"sometimes.dot.option\"]","Operator":"keyin","Comparator":"","Invert":false}'

curl "localhost:8080/blocks/m/set_rule" --data '{"Mask":{"sometimes.dot.option":{},"t":{}}}'

curl "localhost:8080/connect?from=r&to=f"
curl "localhost:8080/connect?from=f&to=m"
curl "localhost:8080/connect?from=m&to=l"
