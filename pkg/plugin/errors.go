package plugin

import "fmt"

// ErrNotSupported indicates that a plugin does not support a particular operation.
//
// Plugins should return this error for optional operations they don't implement.
// For example, a read-only plugin would return this for CreateTask or UpdateTask.
//
// Example:
//
//	func (p *ReadOnlyPlugin) CreateTask(ctx context.Context, ...) (*types.Task, error) {
//	    return nil, plugin.ErrNotSupported
//	}
var ErrNotSupported = fmt.Errorf("operation not supported by this plugin")

// ErrNotConfigured indicates that a plugin has not been properly configured.
//
// Plugins should return this error when ValidateConfig fails or when an operation
// is attempted before Configure has been called.
//
// Use NewErrNotConfigured to provide context about what configuration is missing.
var ErrNotConfigured = fmt.Errorf("plugin not properly configured")

// ErrNotFound indicates that a requested resource does not exist in the external system.
//
// Plugins should return this error when:
//   - FetchProject is called with an unknown project external ID
//   - FetchTask is called with an unknown task external ID
//   - An operation references a non-existent resource
//
// Use NewErrNotFound to provide context about what resource was not found.
var ErrNotFound = fmt.Errorf("resource not found")

// ErrUnauthorized indicates that authentication or authorization failed.
//
// Plugins should return this error when:
//   - API credentials are invalid or expired
//   - The authenticated user doesn't have permission for the operation
//   - Rate limits have been exceeded
//
// Use NewErrUnauthorized to provide context about the authorization failure.
var ErrUnauthorized = fmt.Errorf("unauthorized")

// NewErrNotConfigured creates an ErrNotConfigured error with additional context.
//
// The message should explain what configuration is missing or invalid.
//
// Example:
//
//	if token == "" {
//	    return plugin.NewErrNotConfigured("missing required 'token' configuration")
//	}
func NewErrNotConfigured(message string) error {
	return fmt.Errorf("%w: %s", ErrNotConfigured, message)
}

// NewErrNotFound creates an ErrNotFound error with additional context.
//
// The message should explain what resource was not found.
//
// Example:
//
//	return plugin.NewErrNotFound(fmt.Sprintf("project %s not found", externalID))
func NewErrNotFound(message string) error {
	return fmt.Errorf("%w: %s", ErrNotFound, message)
}

// NewErrUnauthorized creates an ErrUnauthorized error with additional context.
//
// The message should explain why authorization failed.
//
// Example:
//
//	return plugin.NewErrUnauthorized("invalid API token")
func NewErrUnauthorized(message string) error {
	return fmt.Errorf("%w: %s", ErrUnauthorized, message)
}
