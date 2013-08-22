# StreamTools makefile

# requires system environment variable $GOPATH to be set, with a standard GoLang heirarchy, and to be run from the streamtools project directory

# reference
# go get code.google.com/p/go.example/hello
# go install github.com/PuppyKhan/streamtools/import/http_to_streamtools


GOGET=go get
GOINSTALL=go install
GOBIN=$(GOPATH)bin
GOPKGBITLY=$(GOPATH)pkg/darwin_386/github.com/bitly
# NOTE: not sure if or how to abstract "darwin_386"
GOSRCBITLY=$(GOPATH)src/github.com/bitly


GOPKGGOOG=$(GOPATH)pkg/darwin_386/code.google.com/p/go.net
# NOTE: not sure if or how to abstract "darwin_386"
GOSRCGOOG=$(GOPATH)src/code.google.com/p/go.net



# add all executable and library src & targets to BASEPATH
BASEPATH=$(GOBIN):$(GOPKGBITLY):$(GOPKGBITLY)/nsq:$(GOSRCBITLY)/go-simplejson:$(GOSRCBITLY)/nsq/nsq:$(GOPKGGOOG):$(GOSRCGOOG)/websocket
# add BASEPATH and all source directories to VPATH
VPATH=$(BASEPATH):import:import/http_to_streamtools:import/random_stream:export/streamtools_to_pubsub:export/streamtools_to_ws:filter:filter/mask

all: http_to_streamtools random_stream streamtools_to_pubsub streamtools_to_ws mask

# libraries
dependencies: nsq.a go-simplejson.a websocket.a

nsq.a: command_test.go

command_test.go:
	$(GOGET) github.com/bitly/nsq

go-simplejson.a: simplejson.go

simplejson.go:
	$(GOGET) github.com/bitly/go-simplejson

websocket.a: websocket.go

websocket.go:
	$(GOGET) code.google.com/p/go.net/websocket


# streams
http_to_streamtools: nsq.a http_to_streamtools.go
	$(GOINSTALL) ./import/http_to_streamtools

random_stream: go-simplejson.a random_stream.go
	$(GOINSTALL) ./import/random_stream

# csv_to_streamtools: csv_to_streamtools.go
# 	$(GOINSTALL) ./import

# nsq_to_streamtools: nsq_to_streamtools.go
# 	$(GOINSTALL) ./import

streamtools_to_pubsub: nsq.a streamtools_to_pubsub.go
	$(GOINSTALL) ./export/streamtools_to_pubsub

streamtools_to_ws: nsq.a websocket.a streamtools_to_ws.go
	$(GOINSTALL) ./export/streamtools_to_ws

mask: mask.go
	$(GOINSTALL) ./filter/mask

# demux_by_key: demux_by_key.go
# 	$(GOINSTALL) ./filter

# filter_by_keyvalue: filter_by_keyvalue.go
# 	$(GOINSTALL) ./filter

# join_by_key: join_by_key.go
# 	$(GOINSTALL) ./filter




clean:
	rm -f $(GOBIN)/http_to_streamtools
	rm -f $(GOBIN)/random_stream
	rm -f $(GOBIN)/streamtools_to_pubsub
	rm -f $(GOBIN)/streamtools_to_ws
	rm -f $(GOBIN)/mask

