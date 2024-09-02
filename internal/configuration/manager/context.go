package manager

import (
	"context"

	"github.com/TBD54566975/ftl/internal/configuration"
)

type contextKeySecrets struct{}

type contextKeyConfig struct{}

// ContextWithSecrets adds a secrets manager to the given context.
func ContextWithSecrets(ctx context.Context, secretsManager *Manager[configuration.Secrets]) context.Context {
	return context.WithValue(ctx, contextKeySecrets{}, secretsManager)
}

// SecretsFromContext retrieves the secrets configuration.Manager previously
// added to the context with [ContextWithConfig].
func SecretsFromContext(ctx context.Context) *Manager[configuration.Secrets] {
	s, ok := ctx.Value(contextKeySecrets{}).(*Manager[configuration.Secrets])
	if !ok {
		panic("no secrets manager in context")
	}
	return s
}

// ContextWithConfig adds a configuration manager to the given context.
func ContextWithConfig(ctx context.Context, configManager *Manager[configuration.Configuration]) context.Context {
	return context.WithValue(ctx, contextKeyConfig{}, configManager)
}

// ConfigFromContext retrieves the configuration.Manager previously added to the
// context with [ContextWithConfig].
func ConfigFromContext(ctx context.Context) *Manager[configuration.Configuration] {
	m, ok := ctx.Value(contextKeyConfig{}).(*Manager[configuration.Configuration])
	if !ok {
		panic("no configuration manager in context")
	}
	return m
}
