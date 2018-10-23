// Copyright Jetstack Ltd. See LICENSE for details.
package utils

const (
	DefaultLogsPathPlaceholder  = "./[target group]-logs.tar.gz"
	DefaultLogsSincePlaceholder = `$(date --date='24 hours ago')`
	DefaultLogsUntilPlaceholder = `$(date --date='now')`
)
