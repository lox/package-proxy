FROM busybox
MAINTAINER Lachlan Donald <lachlan@ljd.cc>

VOLUME ["/tmp/cache"]
ENTRYPOINT ["/package-proxy"]
CMD ["-dir","/tmp/cache"]
EXPOSE 3142

ADD package-proxy-linux-amd64 /package-proxy
