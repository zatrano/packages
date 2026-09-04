package hashid

import "github.com/zatrano/framework/contracts"

// From resolves the package service from the application container.
func From(app contracts.App) *Hasher {
	return contracts.Resolve[*Hasher](app, "hashid")
}
