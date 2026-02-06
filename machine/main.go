package main

import (
	"github.com/docker/machine/libmachine/drivers/plugin"
	"github.com/Vincent-proginov/docker-machine/machine/driver"
)

func main() {
	plugin.RegisterDriver(driver.NewDriver("", ""))
}
