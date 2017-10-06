// Copyright Jetstack Ltd. See LICENSE for details.
package mocks

// This package contains generated mocks

//go:generate mockgen -imports .=github.com/jetstack/tarmak/pkg/tarmak/interfaces -package=mocks -source=../interfaces/interfaces.go -destination tarmak.go
//go:generate mockgen -package=mocks -source=../provider/amazon/amazon.go -destination amazon.go
