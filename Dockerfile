FROM busybox
MAINTAINER Lachlan Donald <lachlan@ljd.cc>

ADD bin/linux/amd64/package-proxy /package-proxy

VOLUME ["/tmp/cache","/certs"]
ENTRYPOINT ["/package-proxy"]
CMD ["-dir","/tmp/cache"]
EXPOSE 3142 3143