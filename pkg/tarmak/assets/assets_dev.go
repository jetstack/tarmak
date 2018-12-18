// Copyright Jetstack Ltd. See LICENSE for details.

// this package contains generated assets from the repository for use during
// tarmak runtime

// This is only used when building in dev mode - it bundles the wing binary
// +build devmode

package assets

//go:generate go-bindata -prefix ../../../ -pkg $GOPACKAGE -o assets_bindata.go ../../../tagging_control.zip ../../../wing_linux_amd64 ../../../terraform/amazon/modules/... ../../../terraform/amazon/templates/... ../../../puppet/... ../../../packer/...
