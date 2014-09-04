FROM google/golang
WORKDIR /gopath/src/github.com/nytlabs/streamtools
ADD . /gopath/src/github.com/nytlabs/streamtools
RUN make clean
RUN make
RUN ["mkdir", "-p", "/gopath/bin"]
RUN ["ln", "-s", "/gopath/src/github.com/nytlabs/streamtools/build/st", "/gopath/bin/st"]
