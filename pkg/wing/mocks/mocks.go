// Copyright Jetstack Ltd. See LICENSE for details.
package mocks

// This package contains generated mocks

//go:generate mockgen -package=mocks -source=../../../vendor/k8s.io/client-go/rest/request.go -destination http_client.go
//go:generate mockgen -imports .=k8s.io/client-go/rest -package=mocks -source=../../../vendor/k8s.io/client-go/rest/client.go -destination client.go
