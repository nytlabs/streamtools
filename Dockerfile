FROM google/golang
WORKDIR /gopath/src/app
ADD . /gopath/src/app/
ENTRYPOINT /bin/bash
