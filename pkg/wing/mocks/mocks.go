// Copyright Jetstack Ltd. See LICENSE for details.
package mocks

// This package contains generated mocks

//go:generate mockgen -package=mocks -source=../../../vendor/k8s.io/client-go/rest/request.go -destination http_client.go
//go:generate mockgen -package=mocks -source=../command.go -destination command.go
//go:generate mockgen -destination client.go -package=mocks k8s.io/client-go/rest Interface
