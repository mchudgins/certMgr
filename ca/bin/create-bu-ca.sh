#! /bin/bash
#
# initialize business unit ca directory structure
#
# usage:
#   init-bu.sh <unit name> <unit tla>
#
# example:
#   init-bu.sh "Financial Services Group" FSG
#

if [[ $# != 2 ]]; then
  echo "usage:"
  echo "  $0 <business unit name> <unit tla>"
fi

# set the main variables
name=$1
tla=`echo $2 | tr '[:upper:]' '[:lower:]'`
TLA=`echo $2 | tr '[:lower:]' '[:upper:]'`
domain_suffix=dstcorp.io

# make the business unit CA directory, then initialize it
mkdir $tla
cd $tla

if [[ ! -d certs ]]; then
  mkdir certs
fi

if [[ ! -d db ]]; then
  mkdir db
fi

if [[ ! -d private ]]; then
  mkdir private
fi

if [[ ! -d crl ]]; then
  mkdir crl
fi

if [[ -d csr ]]; then
  mkdir csr
fi

chmod 700 private
touch db/index
touch db/index.attr
openssl rand -hex 16 > db/serial
echo 1001 > db/crlnumber

#
# emit the bu-ca.cnf file
#
cat << EOF > bu-ca.cnf

# Intermediate CA configuration file.
# see:  https://www.feistyduck.com/library/openssl-cookbook/online/ch-openssl.html
#
[default]
name              = $tla
domain_suffix     = dstcorp.io
aia_url           = http://$tla.$domain_suffix/$tla.crt
crl_url           = http://$tla.$domain_suffix/$tla.crl
ocsp_url          = http://ocsp.$tla.$domain_suffix:9080
default_ca        = ca_default
name_opt          = utf8,esc_ctrl,multiline,lname,align

[ca_dn]
countryName       = "US"
organizationName  = "DST Systems, Inc"
organizationalUnitName = "DST Internal Use Only -- $TLA Intermediate CA"
commonName = "$tla-ca.$domain_suffix"

[ca_default]
home                    = .
database                = db/index
serial                  = db/serial
crlnumber               = db/crlnumber
certificate             = $tla-ca.crt
private_key             = private/$tla-ca.key
RANDFILE                = private/random
new_certs_dir           = certs
unique_subject          = no
copy_extensions         = none
default_days            = 3650
default_crl_days        = 365
default_md              = sha256
policy                  = policy_c_o_match

[policy_c_o_match]
countryName             = match
stateOrProvinceName     = optional
organizationName        = match
organizationalUnitName  = optional
commonName              = supplied
emailAddress            = optional

[req]
default_bits            = 4096
encrypt_key             = yes
default_md              = sha256
utf8                    = yes
string_mask             = utf8only
prompt                  = no
distinguished_name      = ca_dn
req_extensions          = ca_ext

[ca_ext]
basicConstraints        = critical,CA:true
keyUsage                = critical,keyCertSign,cRLSign
subjectKeyIdentifier    = hash

[sub_ca_ext]
authorityInfoAccess     = @issuer_info
authorityKeyIdentifier  = keyid:always
basicConstraints        = critical,CA:true,pathlen:1
crlDistributionPoints   = @crl_info
extendedKeyUsage        = clientAuth,serverAuth
keyUsage                = critical,keyCertSign,cRLSign
nameConstraints         = @name_constraints
subjectKeyIdentifier    = hash

[crl_info]
URI.0                   = $crl_url

[issuer_info]
caIssuers;URI.0         = $aia_url
OCSP;URI.0              = $ocsp_url

[name_constraints]
permitted;DNS.0=.dstcorp.net
permitted;DNS.1=.dstcorp.io
permitted;DNS.2=.dstcorp.cloud
permitted;DNS.3=.dst.cloud
permitted;DNS.4=.ta2k.com
permitted;DNS.5=.localhost
permitted;DNS.6=.local
permitted;DNS.7=.xip.io
permitted;DNS.8=.internal
#permitted;IP.0=192.168.0.0/255.255.0.0
#permitted;IP.1=172.16.0.0/255.240.0.0
#permitted;IP.2=10.0.0.0/255.0.0.0
#permitted;IP.3=127.0.0.0/255.0.0.0

[ocsp_ext]
authorityKeyIdentifier  = keyid:always
basicConstraints        = critical,CA:false
extendedKeyUsage        = OCSPSigning
keyUsage                = critical,digitalSignature
subjectKeyIdentifier    = hash

EOF

cd -

#Add the necessary extensions for the business unit to the intermediate ca
cd intermediate-ca

cat << EOF > $tla-ca.ext
authorityInfoAccess     = @issuer_info
authorityKeyIdentifier  = keyid:always
basicConstraints        = critical,CA:true,pathlen:0
crlDistributionPoints   = @crl_info
extendedKeyUsage        = clientAuth,serverAuth
keyUsage                = critical,keyCertSign,cRLSign
nameConstraints         = @name_constraints
subjectKeyIdentifier    = hash

[crl_info]
URI.0                   = http://certs.dstcorp.io/fsg.crl

[issuer_info]
caIssuers;URI.0         = http://certs.dstcorp.io/fsg.crt
OCSP;URI.0              = http://ocsp.dstcorp.io:9080

[name_constraints]
permitted;DNS.0=.dstcorp.net
permitted;DNS.1=.dstcorp.io
permitted;DNS.2=.dstcorp.cloud
permitted;DNS.3=.dst.cloud
permitted;DNS.4=.ta2k.com
permitted;DNS.5=.localhost
permitted;DNS.6=.local
permitted;DNS.7=.xip.io
permitted;DNS.8=.internal
#permitted;IP.0=192.168.0.0/255.255.0.0
#permitted;IP.1=172.16.0.0/255.240.0.0
#permitted;IP.2=10.0.0.0/255.0.0.0
#permitted;IP.3=127.0.0.0/255.0.0.0
#excluded;IP.0=0.0.0.0/0.0.0.0
#excluded;IP.1=0:0:0:0:0:0:0:0/0:0:0:0:0:0:0:0

EOF

cd -

#! /bin/bash
#
# generate the business unit's CA
#
cd $tla

if [[ -e private/$tla-ca.key ]]; then
  rm -f private/$tla-ca.key
fi
# create the CSR & key
openssl req -new \
    -config bu-ca.cnf \
    -out $tla-ca.csr \
    -keyout private/$tla-ca.key \
    -passout pass:password
# remove the passphrase from the key
openssl pkey -in private/$tla-ca.key -passin pass:password -out key_unencrypted.pem
rm private/$tla-ca.key
mv key_unencrypted.pem private/$tla-ca.key
chmod 0400 private/$tla-ca.key
cd -

# sign the CSR with the intermediate's CA
cd intermediate-ca
echo -n "signing CSR..."
openssl ca -batch \
  -config intermediate-ca.cnf \
  -extfile $tla-ca.ext \
  -enddate `date -d "15 years" +%Y%m%d120000Z` \
  -notext -md sha256 \
  -in ../$tla/$tla-ca.csr \
  -passin file:private/intermediate-ca.passphrase.bin \
  -out $tla-ca.crt
echo "done"
mv $tla-ca.crt ../$tla/$tla-ca.crt
cd -

echo current directory: `pwd`
# create the CA bundle
echo Creating CA bundle....
cat $tla/$tla-ca.crt > $tla/ca-bundle.pem
cat intermediate-ca/intermediate-ca.crt >> $tla/ca-bundle.pem

echo Done.
