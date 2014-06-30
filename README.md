# Package Proxy

A caching reverse proxy designed for acting as a proxy for package managers. Optionally supports generating self-signed certificates on the fly to allow caching of https resources. 

[![Gobuild Download](http://beta.gobuild.io/badge/github.com/lox/package-proxy/download.png)](http://beta.gobuild.io/github.com/lox/package-proxy)

**Currently supported:**
  * Apt/Ubuntu
  * RubyGems
  * Composer
  * Npm

**Planned**
  * Docker Registry

## Running

Via Docker:

```bash
docker run --tty --interactive --rm --publish 3142:3142 lox24/package-proxy:latest
```

As a binary:

```bash
go get github.com/lox/package-proxy 
$GOBIN/package-proxy
```

By default package-proxy will only run the http proxy, to enable the https proxy:

```bash
$GOBIN/package-proxy -tls
```

## Configuring Package Managers

Where possible, Package Proxy is designed to work as an https/http proxy, so under Linux you should be able to configure it with:

```bash
export http_proxy=http://localhost:3142
export https_proxy=http://localhost:3142
```

Because Package Proxy uses generated SSL certificates (effectively a MITM attack), you need to install the certificate that it generates as a trusted root. **Do not do this unless you understand the security implications**.

**Under Ubuntu:**

```bash
certs/generate-certs.sh
cp packageproxy-ca.crt /usr/local/share/ca-certificates/package-proxy.crt
update-ca-certificates
```

### Apt/Ubuntu

Apt will respect `https_proxy`, but if you'd rather configure it manually

```bash
echo 'Acquire::http::proxy "https://x.x.x.x:3142/";' >> /etc/apt/apt.conf
echo 'Acquire::https::proxy "https://x.x.x.x:3142/";' >> /etc/apt/apt.conf
```

### Development / Releasing

The provided `Dockerfile` will build a development environment. The code will be compiled on every run, so you only need to use `--build` once:

```bash
./docker.sh --build --dev
```

Releasing is a bit complicated as it needs to be built under osx and linux: 

```bash
export GITHUB_TOKEN=xzyxzyxzyxzyxzy 

# under osx
release/release-github.sh v0.6.0 darwin-amd64 package-proxy-darwin-amd64

# now for linux
./docker.sh --build-linux
release/release-github.sh v0.6.0 linux-amd64 package-proxy-linux-amd64

# now the docker image
release/release-docker.sh v0.6.0
```

