// Copyright Jetstack Ltd. See LICENSE for details.

// this package contains generated assets from the repository for use during
// tarmak runtime

// This is only used when building in dev mode - it bundles the wing binary aswell
// +build devmode

package binaries

//go:generate go-bindata -prefix ../../../ -pkg $GOPACKAGE -o binaries_bindata.go ../../../wing_linux_amd64 ../../../tagging_control_linux_amd64
