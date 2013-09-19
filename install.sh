sudo apt-get update
sudo apt-get install -y build-essential
# git
sudo apt-get install -y git
sudo apt-get install -y mercurial
sudo apt-get install -y bzr
# go
wget https://go.googlecode.com/files/go1.1.2.linux-amd64.tar.gz
tar -xzvf go1.1.2.linux-amd64.tar.gz
sudo mv go /usr/local/
# local
mkdir -p ~/go/src/github.com/mikedewar
mkdir -p ~/go/pkg
mkdir -p ~/go/bin
# nsq
wget https://s3.amazonaws.com/bitly-downloads/nsq/nsq-0.2.22.linux-amd64.go1.1.2.tar.gz
tar -xzvf nsq-0.2.22.linux-amd64.go1.1.2.tar.gz 
sudo cp -r nsq-0.2.22.linux-amd64.go1.1.2/* /usr/local/
# paths
touch ~/.bash_profile
echo 'export GOROOT=/usr/local/go' >> ~/.bash_profile
echo 'export GOPATH=$HOME/go' >> ~/.bash_profile
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bash_profile
echo 'export PATH=$PATH:/usr/local/nsq/bin' >> ~/.bash_profile
echo 'export PATH=$PATH:$GOPATH/bin' >> ~/.bash_profile
source ~/.bash_profile
# streamtools dependencies
go get code.google.com/p/snappy-go/snappy
go get launchpad.net/goamz/aws
go get launchpad.net/goamz/ec2
go get github.com/bitly/go-nsq
go get github.com/bitly/go-simplejson
go get github.com/mreiferson/go-snappystream
go get github.com/bmizerany/aws4
# compile streamtools
cd ~/go/src/github.com/mikedewar
git clone git@github.com:mikedewar/stream_tools.git
cd stream_tools/blocks
for dir in `ls`
do 
    cd $dir
    echo $dir
    go install
    cd ..
done


