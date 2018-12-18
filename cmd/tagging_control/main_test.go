// Copyright Jetstack Ltd. See LICENSE for details.
package main

import (
	"crypto/rand"
	"reflect"
	"testing"

	"golang.org/x/crypto/ssh"
)

var (
	RSASigniture = []byte(`L0iuTgj5IfN6XP+VMvuYzuTDRnryuCvKAMcifsECAD5SrauLsB9YF2WUGsmJjBQwVSpVDrCsamLN
TD6ztBNN676pphn3TqeEKzba1EOGo0uQD/ipbKPXjlYYyBe8BLiue3FiDgdnrxLCdwr/vlEXom20
DoiPHfC8YsSPIJXv/ZE=`)

	DocumentRaw = []byte(`{
  "privateIp" : "10.99.32.10",
  "devpayProductCodes" : null,
  "marketplaceProductCodes" : [ "aw0evgkw8e5c1q413zgy5pjce" ],
  "version" : "2017-09-30",
  "instanceType" : "t2.nano",
  "billingProducts" : null,
  "instanceId" : "i-0daab936f4046f7a6",
  "accountId" : "228615251467",
  "availabilityZone" : "eu-west-1a",
  "kernelId" : null,
  "ramdiskId" : null,
  "architecture" : "x86_64",
  "imageId" : "ami-01a7d6d4d6074842c",
  "pendingTime" : "2018-12-18T09:58:51Z",
  "region" : "eu-west-1"
}`)

	RSAPublicKey  = []byte(`ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCgOj0qglXYZEn6QVLMNSVnpuKUCHDqHn88ZSL80ofAqQ3V/8UG/Z2YhZorrM/6gZobPM6P2qIeX1lb2m8zGpcgSXHVVhLDv2ephHTqtX4rPuCHGFdZbWrVcR5ysLy6Jpiy9Yrj+ZuuP/LS7vbZotOXFOSNePPi4dTWDcIOJMcz6z/DMBViXxzcEmUH/8gTK9BI+any68d8SHEpLuZ1HCigYgl6DVvflJ4hW5wVSec80HW7dO5lLjgTs7SRuNejC/09/+TwJk3rWxvcOBTIT1YMG5Y1KyyPth0PEPZtoB1OeZhQoFLhDAx6jb0kzG0OpH3mEd2lIwCof/HyDCkO2XzV`)
	RSAPrivateKey = []byte(`-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEAoDo9KoJV2GRJ+kFSzDUlZ6bilAhw6h5/PGUi/NKHwKkN1f/F
Bv2dmIWaK6zP+oGaGzzOj9qiHl9ZW9pvMxqXIElx1VYSw79nqYR06rV+Kz7ghxhX
WW1q1XEecrC8uiaYsvWK4/mbrj/y0u722aLTlxTkjXjz4uHU1g3CDiTHM+s/wzAV
Yl8c3BJlB//IEyvQSPmp8uvHfEhxKS7mdRwooGIJeg1b35SeIVucFUnnPNB1u3Tu
ZS44E7O0kbjXowv9Pf/k8CZN61sb3DgUyE9WDBuWNSssj7YdDxD2baAdTnmYUKBS
4QwMeo29JMxtDqR95hHdpSMAqH/x8gwpDtl81QIDAQABAoIBACHHJU3o3CAaRF41
lzblnVUUoX+DqAozE6+vwoh5+ZRsDzamDOtEXAzjXXUHoXC2Eb7cOs+oz7SHdVcf
3YFwgZuU4CKRWrNZjoj2G4+/YzHKt5rDTubTYkpM5pZXG/JCYL6ZdQZKgL9jS9Wb
+v42jVS0WtpYPVH/OddGXzqMFlKjWA27tD4wEA23yUVZMQ2ntkXSXvZ0yl4EPtOv
2DoEV+nBEDjhuUD9pNUm7KhMBX71OHNLPmlLezr0G1A2rCPskAl8aV4GOymQSwyT
dkWrdAUwWxbAMKGaUqGFrQA1uWJXrPBjWr7/o4aQTilezwOY3fFvgQKiL1M5L7HP
0/6M3AECgYEAzjhK0Cupd5VQC+ItuOX8Ew0hAKZwEe9bs2RD9s5k4tImCG5A850y
T3Pmpzg6hHhmAXU7sKWTEO3BwK/YMrCVp1UOieFoP3l8q/GOgFFJ/0Bz7MZmnXBc
Lh9JCEssUbibORmbFdhlO12fm8NiQfY8aa4Fg0VbKH2SESypRHPpMaECgYEAxufE
xUC1rZ5eTx70xpAxTAdcDmEEqX3UozyRLp28uLje4GYJcKiJoTEqzlZNiWaugSk/
4VCLNUJrF4bTOPwVnRkvQ71ekmB2JKv8r7gQRuwvcGERzOcdrAMLs1MruVn2Zi5E
68IMT3EOW5r8UeNXTM04CWYYcob5FvgXVZbrprUCgYA5mH9MpOUwAQPaTdF3UsSU
jZYqGFI0sCVsdRSGWh7TOt5kfGano7/pcPV6vrmZRgc3YQbKz3PDxqPWrUY04hzq
H1dwKwRytfucCltCe3GvWNEH0GHYlwkn2JUNO/Gk4Wp5CC3IbCfZ7MwnNOq8gYld
+ryPbU+If4nMQi0EcVswAQKBgQDECMMvIWKtlcsfMcRPOufLJem9pjLhFUoQA+6W
whGxAUtwYEBnj0Pt4TZuHDLY+6F7XPs/hpFc0XQYwOHGZPSsW5jwq1/c5kMqS3OE
f+VS8Q6kNJdFmnbtBCdw+sS6Lgchl/KHZT2awjNDZ5HM50IwSIY1BTGNFqfC0oq0
6UShjQKBgBJ8Ny7gx+8Wwc/aQxsNlMO2v5KCIbhY939uPYAYbav9QXWyWte4sC1+
fDA5+p3yMZdqE/7C/04F2Q4OFayCaNm19MCii9UtpXx8c5Ey53VaXyhhBFXF3xl0
c0vh3QGIdWzlL5RU9UkGIqz1V6CN8UYx4xiOuh95j9L7vv56e8Yw
-----END RSA PRIVATE KEY-----`)

	ECDSAPrivateKey = []byte(`-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIOddZrJaP9vz93byVqS6vCru3lrSgw8nDzGSRvT89BK0oAoGCCqGSM49
AwEHoUQDQgAEjoT2Bd0QKAoVAurwa0xPNUICUnfSJn8mOW6H7F0bGdBGRoTdhvBD
B+fy6asaeXowduiTS+Llunpc6QTdC/YPdA==
-----END EC PRIVATE KEY-----`)
	ECDSAPublicKey = []byte(`ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBI6E9gXdECgKFQLq8GtMTzVCAlJ30iZ/Jjluh+xdGxnQRkaE3YbwQwfn8umrGnl6MHbok0vi5bp6XOkE3Qv2D3Q=`)

	ED25519PrivateKey = []byte(`-----BEGIN OPENSSH PRIVATE KEY-----
b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAMwAAAAtzc2gtZW
QyNTUxOQAAACBadFVUxoD7zee95+3zdw1bTZIL5ZPmXzN4pKUNWH4kSgAAAIiWSm7Wlkpu
1gAAAAtzc2gtZWQyNTUxOQAAACBadFVUxoD7zee95+3zdw1bTZIL5ZPmXzN4pKUNWH4kSg
AAAEDfDWrBPqHSNvN31T00MEMtW4HonMlIoQhZ2q73uk4TMVp0VVTGgPvN573n7fN3DVtN
kgvlk+ZfM3ikpQ1YfiRKAAAAAAECAwQF
-----END OPENSSH PRIVATE KEY-----`)
	ED25519PublicKey = []byte(`ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIFp0VVTGgPvN573n7fN3DVtNkgvlk+ZfM3ikpQ1YfiRK`)
)

func Test_verify(t *testing.T) {
	h := &Handler{
		request: &TagInstanceRequest{
			RSASigniture:        RSASigniture,
			InstanceDocumentRaw: []byte("bad signature"),
			PublicKeys:          make(map[string][]byte),
			KeySignatures:       make(map[string]*ssh.Signature),
		},
	}

	if err := h.verify(); err == nil {
		t.Fatalf("expected error, got=%s", err)
	}

	h.request.InstanceDocumentRaw = DocumentRaw
	if err := h.verify(); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 2; i++ {
		for _, k := range []struct {
			name   string
			sk, pk []byte
		}{
			{"rsa", RSAPrivateKey, RSAPublicKey},
			{"ecdsa", ECDSAPrivateKey, ECDSAPublicKey},
			{"ed25519", ED25519PrivateKey, ED25519PublicKey},
		} {

			signer, err := ssh.ParsePrivateKey(k.sk)
			if err != nil {
				t.Fatal(err)
			}

			sig, err := signer.Sign(rand.Reader, DocumentRaw)
			if err != nil {
				t.Fatal(err)
			}

			h.request.PublicKeys[k.name] = k.pk
			h.request.KeySignatures[k.name] = sig
		}

		err := h.verify()
		if i%2 == 0 {
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
		} else {
			if err == nil {
				t.Fatalf("expecred error, got=%s", err)
			}
		}

		DocumentRaw = []byte("bad signature")
	}
}

func Test_createTags(t *testing.T) {
	h := &Handler{
		request: &TagInstanceRequest{
			PublicKeys: make(map[string][]byte),
		},
	}

	tags := h.createTags()
	checkTags(t, make(map[string]string), tags)

	h.request.PublicKeys = map[string][]byte{
		"key-1": []byte("public key 1"),
		"key-2": []byte("public key 2"),
	}
	tags = h.createTags()
	checkTags(t, map[string]string{
		"tarmak.io/key-1-0": "public key 1==EOF",
		"tarmak.io/key-2-0": "public key 2==EOF",
	}, tags)

	longKey := make([]byte, 256)
	for i := range longKey {
		longKey[i] = 'a'
	}

	h.request.PublicKeys = map[string][]byte{
		"key-1": append(longKey, []byte("bcd")...),
		"key-2": append(longKey, []byte("bcd")...),
	}
	tags = h.createTags()
	checkTags(t, map[string]string{
		"tarmak.io/key-1-0": string(longKey),
		"tarmak.io/key-1-1": "bcd==EOF",
		"tarmak.io/key-2-0": string(longKey),
		"tarmak.io/key-2-1": "bcd==EOF",
	}, tags)
}

func checkTags(t *testing.T, exp, got map[string]string) {
	if !reflect.DeepEqual(exp, got) {
		t.Fatalf("got mismatch of tags\nexp=%s\ngot=%s", exp, got)
	}
}
