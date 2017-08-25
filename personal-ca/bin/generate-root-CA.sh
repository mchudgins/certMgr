#! /bin/bash
#
# generate root CA
#
openssl req -new \
    -config root-ca.cnf \
    -out root-ca.csr \
    -keyout private/root-ca.key \
    -passout file:private/root-ca.passphrase.bin
chmod 0400 private/root-ca.key
openssl ca -batch -selfsign \
    -config root-ca.cnf \
    -in root-ca.csr \
    -keyfile private/root-ca.key \
    -passin file:private/root-ca.passphrase.bin \
    -out root-ca.crt \
    -notext \
    -enddate `date -d "25 years" +%Y%m%d120000Z` \
    -extensions ca_ext
