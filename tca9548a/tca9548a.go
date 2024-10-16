package tca9548a

import (
	"context"
	"encoding/hex"
	"errors"

	"go.viam.com/rdk/components/board/genericlinux/buses"
	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
	"go.viam.com/utils"
)

// Model represents a tca9548a sensor model.
var (
	Model              = resource.NewModel("nfranczak", "sensor", "tca9548a")
	unimplementedError = errors.New("unimplemented")
)

// Non-iota consts
const (
	defaultI2Caddr = 0x2A
)

// AttrConfig is used for converting config attributes.
type Config struct {
	I2CBus  string `json:"i2c_bus"`
	I2CAddr int    `json:"i2c_addr,omitempty"`
}

// Validate ensures all parts of the config are valid.
func (config *Config) Validate(path string) ([]string, error) {
	var deps []string
	if len(config.I2CBus) == 0 {
		return nil, utils.NewConfigValidationFieldRequiredError(path, "i2c_bus")
	}
	return deps, nil
}

func init() {
	resource.RegisterComponent(
		sensor.API,
		Model,
		resource.Registration[sensor.Sensor, *Config]{
			Constructor: func(
				ctx context.Context,
				deps resource.Dependencies,
				conf resource.Config,
				logger logging.Logger,
			) (sensor.Sensor, error) {
				newConf, err := resource.NativeConfig[*Config](conf)
				if err != nil {
					return nil, err
				}
				return newSensor(ctx, deps, conf.ResourceName(), newConf, logger)
			},
		})
}

func newSensor(
	ctx context.Context,
	deps resource.Dependencies,
	name resource.Name,
	attr *Config,
	logger logging.Logger,
) (sensor.Sensor, error) {
	i2cbus, err := buses.NewI2cBus(attr.I2CBus)

	if err != nil {
		return nil, err
	}

	addr := attr.I2CAddr
	if addr == 0 {
		addr = defaultI2Caddr
		logger.Warnf("using i2c address : 0x%s", hex.EncodeToString([]byte{byte(addr)}))
	}

	s := &tca9548a{
		Named:  name.AsNamed(),
		logger: logger,
		bus:    i2cbus,
		addr:   byte(addr),
	}

	return s, nil
}

// tca9548a is a i2c multiplexer that allows interfacing with up to eight nau7802's
type tca9548a struct {
	resource.Named
	resource.AlwaysRebuild
	resource.TriviallyCloseable
	logger logging.Logger

	bus     buses.I2C
	addr    byte
	name    string
	samples int

	zeroOffset        int
	calibrationFactor float64
	gain              int
}

func (s *tca9548a) Readings(ctx context.Context, extra map[string]interface{}) (map[string]interface{}, error) {
	return nil, unimplementedError
}

func (s *tca9548a) DoCommand(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	return nil, unimplementedError
}
