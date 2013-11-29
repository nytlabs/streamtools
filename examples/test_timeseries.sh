./create random
./create timeseries
curl localhost:7070/blocks/2/set_rule --data '{"NumSamples":10, "Key":"random_float"}'
./connect 1 2
