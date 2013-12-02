BLDDIR = build
GUIDIR = gui
DAEMONDIR = daemon
BINARIES = st

all: compiled_html $(BINARIES)
	
$(BLDDIR)/%:
	cd blocks && go get .
	cd daemon && go get .
	go build -o $(BLDDIR)/st ./st

$(BLDDIR)/st: $(wildcard blocks/*.go daemon/*.go st/*.go)

$(BINARIES): %: $(BLDDIR)/%

compiled_html: index.html.go

%.html.go: $(GUIDIR)/%.html
	go get github.com/jteeuwen/go-bindata
	go-bindata -func "index" -pkg "daemon" -out=$(GUIDIR)/index.go ./gui/index.html
	mv $(GUIDIR)/index.go $(DAEMONDIR)

clean: 
	rm -rf $(BLDDIR)
	rm $(DAEMONDIR)/index.go

.PHONY: all
.PHONY: $(BINARIES)
