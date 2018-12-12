// Copyright Jetstack Ltd. See LICENSE for details.
package interfaces

import (
	"github.com/sirupsen/logrus"

	"github.com/jetstack/tarmak/pkg/apis/wing/common"
	client "github.com/jetstack/tarmak/pkg/wing/client/clientset/versioned"
)

type Wing interface {
	Log() *logrus.Entry
	Converge()
	ConvergeWGWait()
	Clientset() *client.Clientset
	Flags() common.Flags
}
