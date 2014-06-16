FROM google/golang

ADD . /gopath/src/github.com/lox/packageproxy
RUN cd /gopath/src/github.com/lox/packageproxy && go get
RUN cd /gopath/src/github.com/lox/packageproxy && go build .

WORKDIR /gopath/src/github.com/lox/packageproxy
EXPOSE 3142
CMD ["./packageproxy","-dir","/tmp/cache"]