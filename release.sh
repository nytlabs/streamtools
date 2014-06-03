#!/bin/bash
VERSION=0.2.6
rm -r release
mkdir release
echo "getting go code"
go get github.com/jteeuwen/go-bindata/...
go-bindata -pkg=server -o st/server/static_bindata.go gui/...
cd st/library && go get .
cd ../server && go get .
cd ..
echo "building"
gox -output="../release/{{.Dir}}_{{.OS}}_{{.Arch}}" -os="linux darwin windows" #-osarch="linux/arm" #
cd ../release
for i in `ls` ; do 
    mkdir tmp; 
    mv $i tmp/st; 
    mv tmp $i-$VERSION/; 
    tar -czf $i-$VERSION.tar.gz $i-$VERSION/; 
    rm -r $i-$VERSION/;
done
