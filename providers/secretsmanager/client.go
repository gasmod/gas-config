package secretsmanager

import (
	"context"
	"errors"
)

// getClient returns the injected client. Building a client from options is
// implemented in a later task.
func (p *Provider) getClient(_ context.Context) (API, error) {
	if p.client != nil {
		return p.client, nil
	}

	return nil, errors.New("no client configured")
}
