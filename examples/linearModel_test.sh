default=7070
PORT=${1:-$default}
echo "Test of the linear model block"
echo ""
echo ""
echo "Make a random block"
curl "localhost:${PORT}/create?blockType=random&id=random"
echo ""
echo ""
echo "Make an linear model block"
curl "localhost:${PORT}/create?blockType=linearModel&id=model"
echo ""
echo ""
echo "Set a rule for the model block"
curl "localhost:${PORT}/blocks/model/set_rule" --data '{"CovariateKey":".['random_float']", "Slope":0.5}'
echo ""
echo ""
curl "localhost:${PORT}/connect?from=random&to=model"
