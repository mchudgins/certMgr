#! /bin/bash
# generate the root, intermediate, and fsg CA's in the current working directory
#
mkdir root-ca intermediate-ca
cd root-ca
../bin/init-root-ca.sh
../bin/generate-root-CA.sh
cd -

cd intermediate-ca
../bin/init-intermediate.sh
../bin/generate-intermediate.sh
cd -

bin/create-bu-ca.sh "k8s" k8s
