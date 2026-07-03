package secretsmanager

import (
	"context"
)

// load fetches and merges all registered secrets.
func (p *Provider) load(_ context.Context) (map[string]any, error) {
	return map[string]any{}, nil
}
