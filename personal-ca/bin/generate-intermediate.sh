#! /bin/bash
#
# generate intermediate CA
#
if [[ -e private/intermediate-ca.key ]]; then
  rm -f private/intermediate-ca.key
fi
openssl req -new \
    -config intermediate-ca.cnf \
    -out intermediate-ca.csr \
    -keyout private/intermediate-ca.key \
    -passout file:private/intermediate-ca.passphrase.bin
chmod 0400 private/intermediate-ca.key
#openssl ca -batch \
#    -config root-ca.cnf \
#    -in root-ca.csr \
#    -passin file:private/intermediate.passphrase.bin \
#    -out root-ca.crt \
#    -extensions ca_ext
cd ../root-ca
openssl ca -batch \
  -config root-ca.cnf \
  -extensions sub_ca_ext \
  -enddate `date -d "20 years" +%Y%m%d120000Z` \
  -notext -md sha256 \
  -in ../intermediate-ca/intermediate-ca.csr \
  -passin file:private/root-ca.passphrase.bin \
  -out intermediate.crt
mv intermediate.crt ../intermediate-ca/intermediate-ca.crt
cd -
