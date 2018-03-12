// Copyright Jetstack Ltd. See LICENSE for details.

// This should only be used when not in devmode
// +build devmode

package assets

//go:generate go-bindata -prefix ../../../ -pkg $GOPACKAGE -o assets_bindata.go ../../../terraform/... ../../../puppet/... ../../../packer/... ../../../tarmak-container_linux_amd64
