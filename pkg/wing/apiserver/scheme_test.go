// Copyright Jetstack Ltd. See LICENSE for details.

package apiserver

import (
	"testing"

	"k8s.io/apimachinery/pkg/api/testing/roundtrip"
)

func TestRoundTripTypes(t *testing.T) {
	roundtrip.RoundTripTestForScheme(t, Scheme, nil)
}
