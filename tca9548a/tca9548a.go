// Package tca9548a implements a tca9548a sensor to get readings from connected peripherals
// We have specifically developed this module to get readings from multiple nau7802's
// datasheet can be found at: https://www.ti.com/lit/ds/symlink/tca9548a.pdf
package tca9548a

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"

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
	defaultI2Caddr = 0x70
)

// AttrConfig is used for converting config attributes.
type Config struct {
	BusName string `json:"bus_name"`
	I2CAddr int    `json:"i2c_addr,omitempty"`
}

// Validate ensures all parts of the config are valid.
func (config *Config) Validate(path string) ([]string, error) {
	var deps []string
	if len(config.BusName) == 0 {
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
	// This is a method that only is available on linux machines
	i2cbus, err := buses.NewI2cBus(attr.BusName)
	if err != nil {
		return nil, err
	}

	addr := attr.I2CAddr
	if addr == 0 {
		addr = defaultI2Caddr
		logger.Warnf("using i2c address : 0x%s", hex.EncodeToString([]byte{byte(addr)}))
	}

	t := &tca9548a{
		Named:  name.AsNamed(),
		logger: logger,
		bus:    i2cbus,
		addr:   byte(addr),
	}

	// THIS MIGHT BE A GOOD IDEA TO JUST SWITCH BETWEEN THE TWO?
	// logger.CDebugf(ctx, "will poll at %d Hz", pollFreq)
	// waitCh := make(chan struct{})
	// pollPerSecond := 1.0 / float64(pollFreq)
	// v.workers = goutils.NewBackgroundStoppableWorkers(func(cancelCtx context.Context) {
	// 	timer := time.NewTicker(time.Duration(pollPerSecond * float64(time.Second)))
	// 	defer timer.Stop()
	// 	close(waitCh)
	// 	for {
	// 		select {
	// 		case <-cancelCtx.Done():
	// 			return
	// 		default:
	// 		}
	// 		select {
	// 		case <-cancelCtx.Done():
	// 			return
	// 		case <-timer.C:
	// 			err := v.getReadings(ctx)
	// 			if err != nil {
	// 				return
	// 			}
	// 		}
	// 	}
	// })
	// <-waitCh

	return t, nil
}

// tca9548a is a i2c multiplexer that allows interfacing with up to eight peripherals
type tca9548a struct {
	resource.Named
	resource.AlwaysRebuild
	resource.TriviallyCloseable
	logger logging.Logger

	bus  buses.I2C
	addr byte
	name string
}

func (t *tca9548a) Readings(ctx context.Context, extra map[string]interface{}) (map[string]interface{}, error) {
	// might need to import the nau7802 and initialize it here before getting readings?
	b, err := t.SelectChannel(ctx, 2)
	if err != nil {
		return nil, err
	}
	data := convertNAU7802Data(b)
	return map[string]interface{}{"adcData": data}, err
}

func (t *tca9548a) SelectChannel(ctx context.Context, channel uint8) ([]byte, error) {
	if channel > 7 {
		return nil, fmt.Errorf("invalid channel: %d", channel)
	}

	// OpenHandle locks the I2C device for communication
	handle, err := t.bus.OpenHandle(t.addr)
	if err != nil {
		return nil, fmt.Errorf("failed to open I2C handle: %w", err)
	}
	defer handle.Close()

	// Write the byte to select the channel (1 << channel)
	data := []byte{1 << channel}
	if err := handle.Write(ctx, data); err != nil {
		return nil, fmt.Errorf("failed to write to I2C: %w", err)
	}

	// Wait for the channel switch to take effect
	// Might need to insert a small delay here, depending on hardware timing requirements.

	// For NAU7802, we expect to read 3 bytes?
	readBuffer := make([]byte, 3)
	if readBuffer, err = handle.Read(ctx, 3); err != nil {
		return nil, fmt.Errorf("failed to read from I2C: %w", err)
	}
	return readBuffer, nil
}

// Convert the 3-byte NAU7802 data to a 24-bit signed integer
func convertNAU7802Data(data []byte) int32 {
	if len(data) != 3 {
		return 0 // Should only handle 3-byte data
	}

	// Combine the bytes to form a 24-bit signed integer
	result := int32(data[0])<<16 | int32(data[1])<<8 | int32(data[2])

	// NAU7802 uses 2's complement for negative numbers
	if result&0x800000 != 0 { // Check if the 24th bit is set (negative number)
		result |= ^0xFFFFFF // Sign extend to 32-bit integer
	}

	return result
}

func (t *tca9548a) DoCommand(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	return nil, unimplementedError
}
