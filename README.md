# Package Proxy

A caching reverse proxy designed for caching package managers. Generates self-signed certificates on the fly to allow caching of https resources.

**Currently supported:**
  * Apt/Ubuntu

**Planned**
  * Composer
  * RubyGems
  * Npm
  * Docker Registry


## Running

Via Docker:

```bash
docker run --tty --interactive --rm --publish 3142:3142 lox/package-proxy:latest      
```

As a binary:

```bash
go get github.com/lox/package-proxy 
$GOBIN/package-proxy
```


## Configuring Package Managers

Where possible, Package Proxy is designed to work as an https/http proxy, so under Linux you should be able to configure it with:

```bash
export HTTP_PROXY=https://localhost:3142
export HTTPS_PROXY=https://localhost:3142
```

Because Package Proxy uses generated SSL certificates (effectively a MITM attack), you need to install the certificate that it generates as a trusted root. **Do not do this unless you understand the security implications**.

```bash
# NOT IMPLEMENTED YET, copy cert from certs dir
sudo mkdir /usr/share/ca-certificates/package-proxy
sudo packageproxy -cert > /usr/share/ca-certificates/package-proxy/package-proxy.crt
sudo dpkg-reconfigure ca-certificates
```


### Apt/Ubuntu

Apt will respect `HTTPS_PROXY`, but if you'd rather configure it manually

```bash
echo 'Acquire::http::proxy "https://x.x.x.x:3142/";' >> /etc/apt/apt.conf
echo 'Acquire::https::proxy "https://x.x.x.x:3142/";' >> /etc/apt/apt.conf
```


