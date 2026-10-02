package openapi

import (
	"github.com/zatrano/framework/v3/core/bootstrap/addons"
)

func init() {
	addons.Register(addons.Meta{
		Name:        "openapi",
		Description: "openapi CLI",
		CLI:         Commands,
	})
}
