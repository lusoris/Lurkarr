# Go Testing with uber/mock (gomock)

> Version: v0.6.0 | License: Apache-2.0
> Docs: https://pkg.go.dev/go.uber.org/mock/gomock
> Mockgen: https://pkg.go.dev/go.uber.org/mock/mockgen

## Overview

uber/mock is the maintained fork of golang/mock. Generates type-safe mock implementations from Go interfaces.

## Installation

```bash
go install go.uber.org/mock/mockgen@latest
```

## Mock Generation (//go:generate)

```go
//go:generate mockgen -destination=mocks/mock_store.go -package=mocks github.com/lusoris/lurkarr/internal/hunting Store
```

Or with source mode:
```go
//go:generate mockgen -source=store.go -destination=mocks/mock_store.go -package=mocks
```

## Usage Pattern

```go
func TestEngine(t *testing.T) {
    ctrl := gomock.NewController(t)
    // No need for ctrl.Finish() — auto-cleanup with testing.T

    mockStore := mocks.NewMockStore(ctrl)

    // Set expectations
    mockStore.EXPECT().
        GetSettings(gomock.Any()).
        Return(&Settings{Enabled: true}, nil).
        Times(1)

    // Create SUT with mock
    engine := hunting.New(mockStore, logger)
    engine.Run(ctx)
}
```

## Matchers

| Matcher | Usage |
|---------|-------|
| `gomock.Any()` | Matches any value |
| `gomock.Eq(x)` | Exact equality |
| `gomock.Nil()` | Nil value |
| `gomock.Not(x)` | Negation |
| `gomock.Len(n)` | Slice/map/string length |
| `gomock.All(m1, m2)` | AND — all matchers must match |
| `gomock.AnyOf(x1, x2)` | OR — any of these values |
| `gomock.InAnyOrder(xs)` | Slice with same elements, any order |
| `gomock.Cond(fn)` | Custom predicate |
| `gomock.Regex(r)` | Regex match on string |
| `gomock.AssignableToTypeOf(x)` | Type assignability |

## Call Expectations

```go
// Exact count
mock.EXPECT().Method(args...).Times(3)

// Range
mock.EXPECT().Method(args...).MinTimes(1).MaxTimes(5)

// Unlimited
mock.EXPECT().Method(args...).AnyTimes()

// Return values
mock.EXPECT().Method(args...).Return(val1, val2)

// Custom action
mock.EXPECT().Method(args...).DoAndReturn(func(x int) (string, error) {
    return fmt.Sprintf("result-%d", x), nil
})

// Call ordering
gomock.InOrder(
    mock.EXPECT().First(),
    mock.EXPECT().Second(),
    mock.EXPECT().Third(),
)
```

## Best Practices for Lurkarr

1. **One mock per interface** — each `Store` interface gets its own mock
2. **Generate mocks in `mocks/` subpackage** or alongside test files
3. **Use `gomock.Any()` for context.Context** — almost always
4. **Test behavior, not implementation** — verify outputs, not internal calls where possible
5. **Table-driven tests** with mocks for comprehensive coverage
6. **Use `DoAndReturn` for complex assertions** on arguments

## Current Lurkarr Mockgen Directives (12)

| Package | Interface | File |
|---------|-----------|------|
| queuecleaner | Store | cleaner.go |
| hunting | ArrHunter | hunter.go |
| hunting | Store | engine.go |
| autoimport | Store | importer.go |
| logging | LogStore | logger.go |
| api | Store | store.go |
| auth | AuthStore | middleware.go |
| cache | settingsLoader | cache.go |
| scheduler | Store | scheduler.go |
| mocks (shared) | scheduler.Store, auth.AuthStore, logging.LogStore | generate.go |

## Coverage Target: 90%+

```bash
# Run with race detection + coverage
go test -race -coverprofile=coverage.out -covermode=atomic ./...

# View HTML report
go tool cover -html=coverage.out -o coverage.html

# Check coverage percentage
go tool cover -func=coverage.out | tail -1
```
