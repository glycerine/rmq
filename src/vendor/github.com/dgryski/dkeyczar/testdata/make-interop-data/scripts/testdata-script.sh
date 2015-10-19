#!/bin/sh -x
#vars for creating test data
PROGLANG="${1?'Missing lang identifier'}" #keyczartool implementation identifier
KEYPROG="${2?'Missing keczar program'}" #how to execute the KeyczarTool
TESTDATA="gen-interop-data/${PROGLANG}_data"

#create symmetic signing key set
$KEYPROG create --location="${TESTDATA}/hmac" --purpose="sign" --name="Test"
$KEYPROG addkey --location="${TESTDATA}/hmac" --status="primary"
$KEYPROG usekey --location="${TESTDATA}/hmac" --destination="${TESTDATA}/hmac/1.out" --format=sign "This is some test data"
$KEYPROG addkey --location="${TESTDATA}/hmac" --status="primary"
$KEYPROG usekey --location="${TESTDATA}/hmac" --destination="${TESTDATA}/hmac/2.out" --format=sign "This is some test data"
$KEYPROG usekey --location="${TESTDATA}/hmac" --destination="${TESTDATA}/hmac/2.timeout" --format=sign-timeout "This is some test data" 2012-12-21T11:11:00Z
$KEYPROG usekey --location="${TESTDATA}/hmac" --destination="${TESTDATA}/hmac/2.unversioned" --format=sign-unversioned "This is some test data"
$KEYPROG usekey --location="${TESTDATA}/hmac" --destination="${TESTDATA}/hmac/2.attached" --format=sign-attached "This is some test data"
$KEYPROG usekey --location="${TESTDATA}/hmac" --destination="${TESTDATA}/hmac/2.secret.attached" --format=sign-attached "This is some test data" "secret"

#create symmetric crypt key set
$KEYPROG create --location="${TESTDATA}/aes" --purpose="crypt" --name="Test"
$KEYPROG addkey --location="${TESTDATA}/aes" --status="primary"
$KEYPROG usekey --location="${TESTDATA}/aes" --destination="${TESTDATA}/aes/1.out" --format=crypt "This is some test data" 
$KEYPROG addkey --location="${TESTDATA}/aes" --status="primary"
$KEYPROG usekey --location="${TESTDATA}/aes"  --destination="${TESTDATA}/aes/2.out" --format=crypt "This is some test data"

#create encrypted symmetric crypting key set
$KEYPROG create --location="${TESTDATA}/aes-crypted" --purpose="crypt" --name="Test"
$KEYPROG addkey --location="${TESTDATA}/aes-crypted" --status="primary" --crypter="${TESTDATA}/aes"
$KEYPROG usekey --location="${TESTDATA}/aes-crypted" --destination="${TESTDATA}/aes-crypted/1.out" --crypter="${TESTDATA}/aes" --format=crypt "This is some test data"
$KEYPROG addkey --location="${TESTDATA}/aes-crypted" --status="primary" --crypter="${TESTDATA}/aes"
$KEYPROG usekey --location="${TESTDATA}/aes-crypted" --destination="${TESTDATA}/aes-crypted/2.out" --crypter="${TESTDATA}/aes" --format=crypt "This is some test data" 

#create symmetric crypting key set without a primary key
$KEYPROG create --location="${TESTDATA}/aes-noprimary" --purpose="crypt" --name="Test" 
$KEYPROG addkey --location="${TESTDATA}/aes-noprimary" --status="primary"
$KEYPROG usekey --location="${TESTDATA}/aes-noprimary" --destination="${TESTDATA}/aes-noprimary/1.out" --format=crypt "This is some test data" 
$KEYPROG demote --location="${TESTDATA}/aes-noprimary" --version="1"

#create symmetric crypt key set various sizes
$KEYPROG create --location="${TESTDATA}/aes-size" --purpose="crypt" --name="Test" 
$KEYPROG addkey --location="${TESTDATA}/aes-size" --status="primary" --size=128
$KEYPROG usekey --location="${TESTDATA}/aes-size" --destination="${TESTDATA}/aes-size/128.out" --format=crypt "This is some test data"
$KEYPROG addkey --location="${TESTDATA}/aes-size" --status="primary" --size=192
$KEYPROG usekey --location="${TESTDATA}/aes-size" --destination="${TESTDATA}/aes-size/192.out" --format=crypt "This is some test data"
$KEYPROG addkey --location="${TESTDATA}/aes-size" --status="primary" --size=256
$KEYPROG usekey --location="${TESTDATA}/aes-size" --destination="${TESTDATA}/aes-size/256.out" --format=crypt "This is some test data"

#create asymmetric private crypting key set
$KEYPROG create --location="${TESTDATA}/rsa" --purpose="crypt" --asymmetric="rsa" --name="Test"
$KEYPROG addkey --location="${TESTDATA}/rsa" --status="primary"
$KEYPROG usekey --location="${TESTDATA}/rsa" --destination="${TESTDATA}/rsa/1.out" --format=crypt "This is some test data"
$KEYPROG addkey --location="${TESTDATA}/rsa" --status="primary"
$KEYPROG usekey --location="${TESTDATA}/rsa" --destination="${TESTDATA}/rsa/2.out" --format=crypt "This is some test data"
#create asymmetric public encrypting key set
$KEYPROG pubkey --location="${TESTDATA}/rsa" --destination="${TESTDATA}/rsa.public"

#create asymmetric private crypting key set rsa various sizes
$KEYPROG create --location="${TESTDATA}/rsa-size" --purpose="crypt" --asymmetric="rsa" --name="Test"
$KEYPROG addkey --location="${TESTDATA}/rsa-size" --status="primary" --size=1024
$KEYPROG usekey --location="${TESTDATA}/rsa-size" --destination="${TESTDATA}/rsa-size/1024.out" --format=crypt "This is some test data"
$KEYPROG addkey --location="${TESTDATA}/rsa-size" --status="primary" --size=2048
$KEYPROG usekey --location="${TESTDATA}/rsa-size" --destination="${TESTDATA}/rsa-size/2048.out" --format=crypt "This is some test data"
$KEYPROG addkey --location="${TESTDATA}/rsa-size" --status="primary" --size=4096
$KEYPROG usekey --location="${TESTDATA}/rsa-size" --destination="${TESTDATA}/rsa-size/4096.out" --format=crypt "This is some test data"
#create asymmetric public encrypting key set  various sizes
$KEYPROG pubkey --location="${TESTDATA}/rsa-size" --destination="${TESTDATA}/rsa-size.public"

#create asymmetric private signing (RSA) keyset 
$KEYPROG create --location="${TESTDATA}/rsa-sign" --purpose="sign" --asymmetric="rsa" --name="Test"
$KEYPROG addkey --location="${TESTDATA}/rsa-sign" --status="primary"
$KEYPROG usekey --location="${TESTDATA}/rsa-sign" --destination="${TESTDATA}/rsa-sign/1.out" --format=sign "This is some test data" 
$KEYPROG addkey --location="${TESTDATA}/rsa-sign" --status="primary"
$KEYPROG usekey --location="${TESTDATA}/rsa-sign" --destination="${TESTDATA}/rsa-sign/2.out" --format=sign "This is some test data" 
$KEYPROG usekey --location="${TESTDATA}/rsa-sign" --destination="${TESTDATA}/rsa-sign/2.timeout" --format=sign-timeout "This is some test data" 2012-12-21T11:11:00Z
$KEYPROG usekey --location="${TESTDATA}/rsa-sign" --destination="${TESTDATA}/rsa-sign/2.unversioned" --format=sign-unversioned "This is some test data"
$KEYPROG usekey --location="${TESTDATA}/rsa-sign" --destination="${TESTDATA}/rsa-sign/2.attached" --format=sign-attached "This is some test data"
$KEYPROG usekey --location="${TESTDATA}/rsa-sign" --destination="${TESTDATA}/rsa-sign/2.secret.attached" --format=sign-attached "This is some test data" "secret"
#create asymmetric public verifying (RSA) keyset
$KEYPROG pubkey --location="${TESTDATA}/rsa-sign" --destination="${TESTDATA}/rsa-sign.public"


#create asymmetric private signing key set rsa various sizes
$KEYPROG create --location="${TESTDATA}/rsa-sign-size" --purpose="sign" --asymmetric="rsa" --name="Test"
$KEYPROG addkey --location="${TESTDATA}/rsa-sign-size" --status="primary" --size=1024
$KEYPROG usekey --location="${TESTDATA}/rsa-sign-size" --destination="${TESTDATA}/rsa-sign-size/1024.out" --format=sign "This is some test data"
$KEYPROG addkey --location="${TESTDATA}/rsa-sign-size" --status="primary" --size=2048
$KEYPROG usekey --location="${TESTDATA}/rsa-sign-size" --destination="${TESTDATA}/rsa-sign-size/2048.out" --format=sign "This is some test data"
$KEYPROG addkey --location="${TESTDATA}/rsa-sign-size" --status="primary" --size=4096
$KEYPROG usekey --location="${TESTDATA}/rsa-sign-size" --destination="${TESTDATA}/rsa-sign-size/4096.out" --format=sign "This is some test data"
#create asymmetric public encrypting key set various sizes
$KEYPROG pubkey --location="${TESTDATA}/rsa-sign-size" --destination="${TESTDATA}/rsa-sign-size.public"

#create asymmetric private signing (DSA) keyset
$KEYPROG create --location="${TESTDATA}/dsa" --purpose="sign" --asymmetric="dsa" --name="Test"
$KEYPROG addkey --location="${TESTDATA}/dsa" --status="primary"
$KEYPROG usekey --location="${TESTDATA}/dsa" --destination="${TESTDATA}/dsa/1.out" --format=sign "This is some test data"
$KEYPROG addkey --location="${TESTDATA}/dsa" --status="primary"
$KEYPROG usekey --location="${TESTDATA}/dsa" --destination="${TESTDATA}/dsa/2.out" --format=sign "This is some test data"
$KEYPROG usekey --location="${TESTDATA}/dsa" --destination="${TESTDATA}/dsa/2.timeout" --format=sign-timeout "This is some test data" 2012-12-21T11:11:00Z
$KEYPROG usekey --location="${TESTDATA}/dsa" --destination="${TESTDATA}/dsa/2.unversioned" --format=sign-unversioned "This is some test data"
$KEYPROG usekey --location="${TESTDATA}/dsa" --destination="${TESTDATA}/dsa/2.attached" --format=sign-attached "This is some test data"
$KEYPROG usekey --location="${TESTDATA}/dsa" --destination="${TESTDATA}/dsa/2.secret.attached" --format=sign-attached "This is some test data" "secret"
#create asymmetric public verifying (DSA) keyset
$KEYPROG pubkey --location="${TESTDATA}/dsa" --destination="${TESTDATA}/dsa.public"

#create crypt session
$KEYPROG usekey --location="${TESTDATA}/rsa.public" --destination="${TESTDATA}/rsa/2.session.material" --destination2="${TESTDATA}/rsa/2.session.ciphertext" --format=crypt-session "This is some test data"
#create crypt signed session
$KEYPROG usekey --location="${TESTDATA}/rsa.public" --location2="${TESTDATA}/dsa" --destination="${TESTDATA}/rsa/2.signedsession.material" --destination2="${TESTDATA}/rsa/2.signedsession.ciphertext" --format=crypt-signedsession "This is some test data"

