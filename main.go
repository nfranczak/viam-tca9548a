// package main is a module providing a tca9548a driver
package main

import (
	"context"

	// tca9548a "viam-TCA9548A/tca9548a"

	// "github.com/nfranczak/viam-tca9548a/tca9548a"
	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/module"
	"go.viam.com/utils"
	"main.go/tca9548a"
)

func main() {
	utils.ContextualMain(mainWithArgs, logging.NewLogger("tca9548a"))
}

func mainWithArgs(ctx context.Context, args []string, logger logging.Logger) error {
	tca9548aModule, err := module.NewModuleFromArgs(ctx, logger)
	if err != nil {
		return err
	}

	tca9548aModule.AddModelFromRegistry(ctx, sensor.API, tca9548a.Model)

	err = tca9548aModule.Start(ctx)
	defer tca9548aModule.Close(ctx)
	if err != nil {
		return err
	}

	<-ctx.Done()
	return nil
}
