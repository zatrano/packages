package factory

import (
	"github.com/zatrano/framework/v2/bootstrap/addons"
)

func init() {
	addons.Register(addons.Meta{
		Name:        "factory",
		Description: "factory CLI",
		CLI:         Commands,
	})
}
