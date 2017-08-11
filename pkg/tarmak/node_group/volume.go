package node_group

import (
	"fmt"

	"github.com/jetstack/tarmak/pkg/tarmak/config"
	"github.com/jetstack/tarmak/pkg/tarmak/interfaces"
)

var _ interfaces.Volume = &Volume{}

type Volume struct {
	conf *config.Volume

	device string
}

func NewVolumeFromConfig(pos int, conf *config.Volume) (*Volume, error) {
	volume := &Volume{
		conf: conf,
	}

	if conf.AWS != nil && pos < 10 {
		letters := "defghijklmnop"
		volume.device = fmt.Sprintf("/dev/sd%r", letters[pos])
	}

	return volume, nil
}

func (v *Volume) Name() string {
	return v.conf.Name
}

func (v *Volume) Size() int {
	return v.conf.Size
}

func (v *Volume) Type() string {
	if v.conf.AWS != nil {
		return v.conf.AWS.Type
	}
	return ""
}
