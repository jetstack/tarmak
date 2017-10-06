// Copyright Jetstack Ltd. See LICENSE for details.
package amazon

import (
	"errors"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/golang/mock/gomock"
	"golang.org/x/crypto/ssh"
)

var fakeSSHKeyInsecurePrivate = `-----BEGIN RSA PRIVATE KEY-----
MIIEogIBAAKCAQEA6NF8iallvQVp22WDkTkyrtvp9eWW6A8YVr+kz4TjGYe7gHzI
w+niNltGEFHzD8+v1I2YJ6oXevct1YeS0o9HZyN1Q9qgCgzUFtdOKLv6IedplqoP
kcmF0aYet2PkEDo3MlTBckFXPITAMzF8dJSIFo9D8HfdOV0IAdx4O7PtixWKn5y2
hMNG0zQPyUecp4pzC6kivAIhyfHilFR61RGL+GPXQ2MWZWFYbAGjyiYJnAmCP3NO
Td0jMZEnDkbUvxhMmBYSdETk1rRgm+R4LOzFUGaHqHDLKLX+FIPKcF96hrucXzcW
yLbIbEgE98OHlnVYCzRdK8jlqm8tehUc9c9WhQIBIwKCAQEA4iqWPJXtzZA68mKd
ELs4jJsdyky+ewdZeNds5tjcnHU5zUYE25K+ffJED9qUWICcLZDc81TGWjHyAqD1
Bw7XpgUwFgeUJwUlzQurAv+/ySnxiwuaGJfhFM1CaQHzfXphgVml+fZUvnJUTvzf
TK2Lg6EdbUE9TarUlBf/xPfuEhMSlIE5keb/Zz3/LUlRg8yDqz5w+QWVJ4utnKnK
iqwZN0mwpwU7YSyJhlT4YV1F3n4YjLswM5wJs2oqm0jssQu/BT0tyEXNDYBLEF4A
sClaWuSJ2kjq7KhrrYXzagqhnSei9ODYFShJu8UWVec3Ihb5ZXlzO6vdNQ1J9Xsf
4m+2ywKBgQD6qFxx/Rv9CNN96l/4rb14HKirC2o/orApiHmHDsURs5rUKDx0f9iP
cXN7S1uePXuJRK/5hsubaOCx3Owd2u9gD6Oq0CsMkE4CUSiJcYrMANtx54cGH7Rk
EjFZxK8xAv1ldELEyxrFqkbE4BKd8QOt414qjvTGyAK+OLD3M2QdCQKBgQDtx8pN
CAxR7yhHbIWT1AH66+XWN8bXq7l3RO/ukeaci98JfkbkxURZhtxV/HHuvUhnPLdX
3TwygPBYZFNo4pzVEhzWoTtnEtrFueKxyc3+LjZpuo+mBlQ6ORtfgkr9gBVphXZG
YEzkCD3lVdl8L4cw9BVpKrJCs1c5taGjDgdInQKBgHm/fVvv96bJxc9x1tffXAcj
3OVdUN0UgXNCSaf/3A/phbeBQe9xS+3mpc4r6qvx+iy69mNBeNZ0xOitIjpjBo2+
dBEjSBwLk5q5tJqHmy/jKMJL4n9ROlx93XS+njxgibTvU6Fp9w+NOFD/HvxB3Tcz
6+jJF85D5BNAG3DBMKBjAoGBAOAxZvgsKN+JuENXsST7F89Tck2iTcQIT8g5rwWC
P9Vt74yboe2kDT531w8+egz7nAmRBKNM751U/95P9t88EDacDI/Z2OwnuFQHCPDF
llYOUI+SpLJ6/vURRbHSnnn8a/XG+nzedGH5JGqEJNQsz+xT2axM0/W/CRknmGaJ
kda/AoGANWrLCz708y7VYgAtW2Uf1DPOIYMdvo6fxIB5i9ZfISgcJ/bbCUkFrhoH
+vq/5CIWxCPp0f85R4qxxQ5ihxJ0YDQT9Jpx4TMss4PSavPaBH3RXow5Ohe+bYoQ
NE5OgEXk2wVfZczCZpigBKbKZHNYcelXtTt/nP3rsCuGcM4h53s=
-----END RSA PRIVATE KEY-----
`

var fakeSSHKeyInsecureFingerprint = "c7:15:68:10:e9:39:6c:ab:99:fe:d0:8b:e8:ec:f5:ae"

var fakeSSHKeyInsecurePublic = "ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEA6NF8iallvQVp22WDkTkyrtvp9eWW6A8YVr+kz4TjGYe7gHzIw+niNltGEFHzD8+v1I2YJ6oXevct1YeS0o9HZyN1Q9qgCgzUFtdOKLv6IedplqoPkcmF0aYet2PkEDo3MlTBckFXPITAMzF8dJSIFo9D8HfdOV0IAdx4O7PtixWKn5y2hMNG0zQPyUecp4pzC6kivAIhyfHilFR61RGL+GPXQ2MWZWFYbAGjyiYJnAmCP3NOTd0jMZEnDkbUvxhMmBYSdETk1rRgm+R4LOzFUGaHqHDLKLX+FIPKcF96hrucXzcWyLbIbEgE98OHlnVYCzRdK8jlqm8tehUc9c9WhQ==\n"

// test the happy path, local key matches the one existing in Amazon
func TestAmazon_validateAmazonKeyPairExistingHappyPath(t *testing.T) {
	a := newFakeAmazon(t)
	defer a.ctrl.Finish()

	// amazon repsonds with one key
	a.fakeEC2.EXPECT().DescribeKeyPairs(gomock.Any()).Return(
		&ec2.DescribeKeyPairsOutput{
			KeyPairs: []*ec2.KeyPairInfo{
				&ec2.KeyPairInfo{
					KeyFingerprint: aws.String(fakeSSHKeyInsecureFingerprint),
					KeyName:        aws.String("myfake_key"),
				},
			},
		},
		nil,
	)

	// environment provider the right public key
	signer, err := ssh.ParseRawPrivateKey([]byte(fakeSSHKeyInsecurePrivate))
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	a.fakeEnvironment.EXPECT().SSHPrivateKey().Return(signer)

	err = a.Amazon.validateAWSKeyPair()
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestAmazon_validateAmazonKeyPairExistingMismatch(t *testing.T) {
	a := newFakeAmazon(t)
	defer a.ctrl.Finish()

	// amazon repsonds with one key
	a.fakeEC2.EXPECT().DescribeKeyPairs(gomock.Any()).Return(
		&ec2.DescribeKeyPairsOutput{
			KeyPairs: []*ec2.KeyPairInfo{
				&ec2.KeyPairInfo{
					KeyFingerprint: aws.String("c7:15:68:10:e9:39:6c:ab:99:fe:d0:8b:e8:ec:f5:xx"),
					KeyName:        aws.String("myfake_key"),
				},
			},
		},
		nil,
	)

	// environment provider the right public key
	signer, err := ssh.ParseRawPrivateKey([]byte(fakeSSHKeyInsecurePrivate))
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	a.fakeEnvironment.EXPECT().SSHPrivateKey().Return(signer)

	err = a.Amazon.validateAWSKeyPair()
	if err == nil {
		t.Errorf("expected an error: %s", err)
	} else if !strings.Contains(err.Error(), "key pair does not match") {
		t.Errorf("unexpected error message: %s", err)
	}
}

func TestAmazon_validateAmazonKeyPairNotExisting(t *testing.T) {
	a := newFakeAmazon(t)
	defer a.ctrl.Finish()

	// amazon reports no key
	a.fakeEC2.EXPECT().DescribeKeyPairs(gomock.Any()).Return(
		&ec2.DescribeKeyPairsOutput{
			KeyPairs: []*ec2.KeyPairInfo{},
		},
		awserr.New("InvalidKeyPair.NotFound", "keypair not found", errors.New("not found")),
	)

	// aws get the import command
	a.fakeEC2.EXPECT().ImportKeyPair(&ec2.ImportKeyPairInput{
		KeyName:           aws.String("myfake_key"),
		PublicKeyMaterial: []byte(fakeSSHKeyInsecurePublic),
	}).Return(
		&ec2.ImportKeyPairOutput{
			KeyFingerprint: aws.String(fakeSSHKeyInsecureFingerprint),
			KeyName:        aws.String("myfake_key"),
		},
		nil,
	)

	// environment provider the right public key
	signer, err := ssh.ParseRawPrivateKey([]byte(fakeSSHKeyInsecurePrivate))
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	a.fakeEnvironment.EXPECT().SSHPrivateKey().Return(signer)

	err = a.Amazon.validateAWSKeyPair()
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}
