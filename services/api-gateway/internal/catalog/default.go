//go:build !postext

package catalog

import "api-gateway/internal/clients"

// FromClients returns no capabilities: the Post and Connection services do not
// yet expose the RPCs they would need. Build with -tags postext once they do.
func FromClients(clients.Set) (PostCatalog, Recommender) { return nil, nil }
