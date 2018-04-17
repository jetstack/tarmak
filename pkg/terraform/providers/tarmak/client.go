// Copyright Jetstack Ltd. See LICENSE for details.
package tarmak

import (
	"fmt"
	"log"
	"net/rpc"
	"sync"
	"time"
)

var rpcClient *rpc.Client
var rpcClientLock sync.Mutex

// ensure there is exactly one RPC client connection
func newRPCClient(socketPath string) (conn *rpc.Client, err error) {
	rpcClientLock.Lock()
	defer rpcClientLock.Unlock()
	if rpcClient == nil {
		tries := 20
		for {
			log.Printf("[DEBUG] trying to connect to tarmak using unix socket '%s'", socketPath)
			conn, err := rpc.Dial("unix", socketPath)
			rpcClient = conn
			if err == nil {
				break
			}
			if err != nil {
				log.Printf("[WARN] unable to dial into unix socket '%s': %v\n", socketPath, err)
			}
			if tries == 0 {
				return nil, fmt.Errorf("error connecting to connector socket '%s': %s", socketPath, err)
			}
			tries -= 1
			time.Sleep(time.Second)
		}
	}

	return rpcClient, nil
}
