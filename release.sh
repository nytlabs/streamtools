#!/bin/bash
go get github.com/jteeuwen/go-bindata/...
go-bindata -pkg=server -o st/server/static_bindata.go gui/...
cd st/library && go get .
cd ../server && go get .
cd ..
gox -output="../build/{{.Dir}}_{{.OS}}_{{.Arch}}" -os="linux darwin windows"
