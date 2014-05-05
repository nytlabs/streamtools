# ./docs/documentup.sh ~/Projects/labs/streamtools/docs/index.html
# you tell this the path to your streamtools gh-pages checkout
# posts documentation markdown to documentup.com, outputs to $destination
curl -X POST --data-urlencode name=streamtools --data-urlencode content@docs/docs.md http://documentup.com/compiled > $1
