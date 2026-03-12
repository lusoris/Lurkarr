// Package mocks provides generated mock implementations for cross-package testing.
package mocks

//go:generate mockgen -typed -destination mock_scheduler_store.go -package mocks github.com/lusoris/lurkarr/internal/scheduler Store
//go:generate mockgen -typed -destination mock_auth_store.go -package mocks github.com/lusoris/lurkarr/internal/auth AuthStore
