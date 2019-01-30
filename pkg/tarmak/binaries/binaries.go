// Copyright Jetstack Ltd. See LICENSE for details.

// this package contains binaries of wing and tagging_control for use during
// tarmak runtime

package binaries

//go:generate go-bindata -prefix ../../../_output -pkg $GOPACKAGE -o binaries_bindata.go ../../../_output/tagging_control_linux_amd64 ../../../_output/wing_linux_amd64
