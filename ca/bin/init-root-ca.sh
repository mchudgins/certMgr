#! /bin/bash
#
# initialize root ca directory structure
#
# generate a random 128 byte passphrase:
#   curl -s https://www.fourmilab.ch/cgi-bin/Hotbits?nbytes=128\&fmt=bin >fubar.bin
#
# remove a passphrase from a private key
#   openssl rsa -in key.pem -out key_unencrypted.pem
#
if [[ ! -d certs ]]; then
  mkdir certs
fi

if [[ ! -d db ]]; then
  mkdir db
fi

if [[ ! -d private ]]; then
  mkdir private
fi

if [[ ! -e private/root-ca.passphrase.bin ]]; then
  curl -s https://www.fourmilab.ch/cgi-bin/Hotbits?nbytes=128\&fmt=bin >private/root-ca.passphrase.bin
  chmod 0400 private/root-ca.passphrase.bin
fi

chmod 700 private
touch db/index
openssl rand -hex 16 > db/serial
echo 1001 > db/crlnumber

#
# emit the root-ca.cnf file
#
cat << 'EOF' > root-ca.cnf

# Root CA configuration file.
# see:  https://www.feistyduck.com/library/openssl-cookbook/online/ch-openssl.html
#
[default]
name              = root-ca
domain_suffix     = dstcorp.io
aia_url           = http://$name.$domain_suffix/$name.crt
crl_url           = http://$name.$domain_suffix/$name.crl
ocsp_url          = http://ocsp.$name.$domain_suffix:9080
default_ca        = ca_default
name_opt          = utf8,esc_ctrl,multiline,lname,align

[ca_dn]
countryName       = "US"
organizationName  = "DST Systems, Inc"
commonName        = "DST Internal Use Only -- ROOT CA"

[ca_default]
home                    = .
database                = $home/db/index
serial                  = $home/db/serial
crlnumber               = $home/db/crlnumber
certificate             = $home/$name.crt
private_key             = $home/private/$name.key
RANDFILE                = $home/private/random
new_certs_dir           = $home/certs
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
nameConstraints         = @name_constraints

[sub_ca_ext]
authorityInfoAccess     = @issuer_info
authorityKeyIdentifier  = keyid:always
basicConstraints        = critical,CA:true,pathlen:2
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
permitted;DNS.0=dstcorp.net
permitted;DNS.1=awddev.dstcorp.net
permitted;DNS.2=dstcorp.io
permitted;DNS.3=dstcorp.cloud
permitted;DNS.4=ta2k.com
permitted;DNS.5=test
permitted;DNS.6=localhost
permitted;IP.0=192.168.0.0/255.255.0.0
permitted;IP.1=172.16.0.0/255.240.0.0
permitted;IP.2=10.0.0.0/255.0.0.0
#excluded;IP.0=0.0.0.0/0.0.0.0
#excluded;IP.1=0:0:0:0:0:0:0:0/0:0:0:0:0:0:0:0

[ocsp_ext]
authorityKeyIdentifier  = keyid:always
basicConstraints        = critical,CA:false
extendedKeyUsage        = OCSPSigning
keyUsage                = critical,digitalSignature
subjectKeyIdentifier    = hash

EOF
