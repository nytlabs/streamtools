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
	go-bindata -func "index" -pkg "daemon" -out=$(GUIDIR)/static_index.go ./gui/index.html
	go-bindata -func "js_main" -pkg "daemon" -out=$(GUIDIR)/static_js_main.go ./gui/static/main.js
	go-bindata -func "lib_d3" -pkg "daemon" -out=$(GUIDIR)/static_lib_d3.go ./gui/static/d3.v3.min.js
	go-bindata -func "lib_jquery" -pkg "daemon" -out=$(GUIDIR)/static_lib_jquery.go ./gui/static/jquery-2.1.0.min.js
	go-bindata -func "lib_underscore" -pkg "daemon" -out=$(GUIDIR)/static_lib_underscore.go ./gui/static/underscore-min.js	
	go-bindata -func "css_main" -pkg "daemon" -out=$(GUIDIR)/static_css_main.go ./gui/static/main.css

	mv $(GUIDIR)/static_index.go $(DAEMONDIR)
	mv $(GUIDIR)/static_js_main.go $(DAEMONDIR)
	mv $(GUIDIR)/static_lib_d3.go $(DAEMONDIR)
	mv $(GUIDIR)/static_lib_jquery.go $(DAEMONDIR)
	mv $(GUIDIR)/static_lib_underscore.go $(DAEMONDIR)
	mv $(GUIDIR)/static_css_main.go $(DAEMONDIR)

clean: 
	rm -rf $(BLDDIR)
	rm $(DAEMONDIR)/_*.go

.PHONY: all
.PHONY: $(BINARIES)
