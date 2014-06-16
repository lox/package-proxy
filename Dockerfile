FROM busybox
MAINTAINER Lachlan Donald <lachlan@ljd.cc>

ADD package-proxy-linux-amd64 /package-proxy
VOLUME ["/tmp/cache"]

CMD ["/package-proxy","-dir","/tmp/cache"]
EXPOSE 3142
