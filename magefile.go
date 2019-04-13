//+build mage

package main

import "github.com/magefile/mage/sh"

// Generate GraphQL models and resolvers
func GenerateGraphQL() error {
	return sh.Run("go", "run", "github.com/99designs/gqlgen")
}

// Run the GraphQL server
func ServeGraphQL() error {
	return sh.Run("go", "run", "server/server.go")
}
