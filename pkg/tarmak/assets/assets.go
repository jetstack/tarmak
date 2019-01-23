// Copyright Jetstack Ltd. See LICENSE for details.

// this package contains generated assets from the repository for use during
// tarmak runtime

// This should be used when running in release mode (!devmode)
// +build !devmode

package assets

//go:generate go run ../../../cmd/tagging_control/main.go zip ../../../tagging_control.zip ../../../tagging_control_linux_amd64
//go:generate go-bindata -prefix ../../../ -pkg $GOPACKAGE -o assets_bindata.go ../../../tagging_control.zip ../../../terraform/amazon/modules/... ../../../terraform/amazon/templates/... ../../../puppet/... ../../../packer/...
