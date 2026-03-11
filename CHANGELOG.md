# Changelog

## 1.0.0 (2026-03-11)


### Features

* full rewrite as Lurkarr (Go + SvelteKit) ([9551d29](https://github.com/lusoris/Lurkarr/commit/9551d29ed344e07cce4f4ed339b4722de939ae9b))
* queue management, auto-import, scoring dedup, security fixes ([34cfc93](https://github.com/lusoris/Lurkarr/commit/34cfc93039da078b0e5f2cd79a683d85d33063d3))


### Bug Fixes

* **ci:** bump Go 1.25.8, golangci-lint v2.11.3, Dockerfile pin ([770b404](https://github.com/lusoris/Lurkarr/commit/770b404cac80a102369d90d7131d793ffd2b43f3))
* **ci:** go mod tidy, bump Go 1.25.3, golangci-lint v2, pin Dockerfile ([0e09e41](https://github.com/lusoris/Lurkarr/commit/0e09e41d747a434809dd29eae0dcd5e77452ed18))
* **ci:** gofmt as formatter in v2, Trivy non-blocking ([b825ad4](https://github.com/lusoris/Lurkarr/commit/b825ad46ff4f600f61c4a90f442478935a3df60e))
* **ci:** golangci-lint v2 config format, Trivy ignore-unfixed ([b5b5718](https://github.com/lusoris/Lurkarr/commit/b5b5718eda925922d44a9e488b27fcebf994da74))
* **ci:** golangci-lint v2 exclusions under linters, scratch base image ([8fb1a60](https://github.com/lusoris/Lurkarr/commit/8fb1a6019a0ff1ae0e7d4c40a8987ce964867906))
* **ci:** remove gosimple (merged into staticcheck in v2) ([1de6fe6](https://github.com/lusoris/Lurkarr/commit/1de6fe652f9277ff103ff0b8848481da28c4ed25))
* nolint for intentional nilerr in totp, fix gofmt in ratelimit ([0fb152e](https://github.com/lusoris/Lurkarr/commit/0fb152e3b8d9442408a4ac5ed9500779d9607294))
* resolve all 27 golangci-lint v2 code errors ([c289fef](https://github.com/lusoris/Lurkarr/commit/c289fef908b5312f57edc9a41ce61b2f013d0b49))
* resolve remaining lint errors in totp, sabnzbd, config, ratelimit ([8ebe108](https://github.com/lusoris/Lurkarr/commit/8ebe108f4153ab36cfcd481aa141fe43d2d8736c))
