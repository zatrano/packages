package oauth

import "github.com/zatrano/framework/v2/contracts"

// From resolves the OAuth server from the application container.
func From(app contracts.App) *Server {
	return contracts.Resolve[*Server](app, "oauth")
}
