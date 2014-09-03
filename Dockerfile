FROM google/golang
WORKDIR /gopath/src/github.com/nytlabs/streamtools
ADD . /gopath/src/github.com/nytlabs/streamtools
RUN make clean
RUN make
