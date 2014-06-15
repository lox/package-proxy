#!/bin/bash

CERTDIR=`pwd`

# generate a private key
openssl genrsa 4096 > $CERTDIR/private.key

# generate a certificate in pem format
echo -e "AU\nVictoria\nMelbourne\nPackage Manager Proxy\nPackage Manager Proxy\n*\nlachlan@ljd.cc\n" |\
  openssl req -new -x509 -days 3650 -key $CERTDIR/private.key -out $CERTDIR/public.pem