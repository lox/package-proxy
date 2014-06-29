#!/bin/bash

cd $(dirname $0)/certs

# generate a private key
openssl genrsa 4096 > packageproxy-ca.key

# generate a certificate in pem format
openssl req -new -x509 -days 120 -key packageproxy-ca.key -out packageproxy-ca.crt \
  -subj "/C=AU/ST=Victoria/O=Fake Organization/CN=Package Proxy CA"
