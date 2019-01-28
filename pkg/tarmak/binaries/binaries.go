// Copyright Jetstack Ltd. See LICENSE for details.

// this package contains binaries of wing and tagging_control for use during
// tarmak runtime

package binaries

//go:generate go-bindata -prefix ../../../ -pkg $GOPACKAGE -o binaries_bindata.go ../../../cmd/tagging_control/tagging_control_linux_amd64 ../../../cmd/wing/wing_linux_amd64
