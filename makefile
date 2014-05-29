BLDDIR = build
GUIDIR = gui
SERVERDIR = st/server
BINARIES = st

all: $(BINARIES)

$(BLDDIR)/%:
	go get github.com/jteeuwen/go-bindata/...
	go-bindata -pkg=server -o st/server/static_bindata.go gui/... examples/...
	go get github.com/tools/godep/...
	godep restore ./...
	go build -o $(BLDDIR)/st ./st

$(BLDDIR)/st: $(wildcard blocks/*.go $(SERVERDIR)/*.go st/*.go)

$(BINARIES): %: $(BLDDIR)/%

clean: 
	rm -rf $(BLDDIR)
	rm $(SERVERDIR)/static_bindata.go


.PHONY: all
.PHONY: $(BINARIES)
