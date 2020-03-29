#!/bin/bash

ROOT_PATH=$(pwd)
EASYRSA_VERSION=$1
CA_PATH=EasyRSA-${EASYRSA_VERSION}/pki/
CERTS_PATH=EasyRSA-${EASYRSA_VERSION}/pki/issued
KEYS_PATH=EasyRSA-${EASYRSA_VERSION}/pki/private

wget https://github.com/OpenVPN/easy-rsa/releases/download/${EASYRSA_VERSION}/EasyRSA-unix-${EASYRSA_VERSION}.tgz -O - | tar -xz
cd EasyRSA-${EASYRSA_VERSION}
./easyrsa init-pki
EASYRSA_BATCH=1 ./easyrsa build-ca nopass
./easyrsa build-server-full marin3r-server nopass
./easyrsa build-client-full envoy-client nopass
./easyrsa build-client-full envoy-server nopass

cd ${ROOT_PATH}
mkdir -p certs

cp ${CA_PATH}/ca.crt certs/ca.crt
cp ${KEYS_PATH}/* certs/
echo -e "$(openssl x509 -inform pem -in ${CERTS_PATH}/envoy-client.crt)" > certs/envoy-client.crt
echo -e "$(openssl x509 -inform pem -in ${CERTS_PATH}/envoy-server.crt)" > certs/envoy-server.crt
echo -e "$(openssl x509 -inform pem -in ${CERTS_PATH}/marin3r-server.crt)" > certs/marin3r-server.crt

rm -rf EasyRSA-${EASYRSA_VERSION}