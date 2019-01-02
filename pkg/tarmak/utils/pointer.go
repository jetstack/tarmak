// Copyright Jetstack Ltd. See LICENSE for details.
package utils

func PointerInt32(i int) *int32 {
	j := int32(i)
	return &j
}
