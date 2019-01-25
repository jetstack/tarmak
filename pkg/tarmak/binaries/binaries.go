// Copyright Jetstack Ltd. See LICENSE for details.

// this package contains generated assets from the repository for use during
// tarmak runtime

// This should be used when running in release mode (!devmode)
// +build !devmode

package binaries

//go:generate go-bindata -prefix ../../../ -pkg $GOPACKAGE -o binaries_bindata.go ../../../tagging_control_linux_amd64
