package common

// Typed wrappers to avoid string collision in the DI container.
// samber/do resolves by concrete type, so WorkDir and DrupalRoot
// would conflict as plain strings.

type WorkDir string

type DrupalRoot string
