BLDDIR = build
GUIDIR = gui
DAEMONDIR = daemon
BINARIES = st

all: $(BINARIES)

$(BLDDIR)/%:
	go get github.com/jteeuwen/go-bindata/...
	go-bindata -pkg=daemon -o daemon/static_bindata.go gui/...
	cd blocks && go get .
	cd daemon && go get .
	go build -o $(BLDDIR)/st ./st

$(BLDDIR)/st: $(wildcard blocks/*.go daemon/*.go st/*.go)

$(BINARIES): %: $(BLDDIR)/%

clean: 
	rm -rf $(BLDDIR)
	rm $(DAEMONDIR)/static_bindata.go


.PHONY: all
.PHONY: $(BINARIES)
