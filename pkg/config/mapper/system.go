package config_mapper

import (
	"fmt"

	"github.com/whiteboxio/flow/pkg/config"
)

func GetSystemCfg() (*config.CfgBlockSystem, error) {
	vIntf, ok := config.Get("general.system")
	if !ok {
		return nil, fmt.Errorf("Missing system config block")
	}
	v, convOk := vIntf.(config.CfgBlockSystem)
	if !convOk {
		return nil, fmt.Errorf(
			"Malformed system config type: expected: config_mappers.CfgBlockSystem, got: %+v", vIntf)
	}
	return &v, nil
}
