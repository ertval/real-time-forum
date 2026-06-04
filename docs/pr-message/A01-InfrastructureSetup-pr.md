# PR Message - A01: Infrastructure, CI/CD, and Dev Tools Setup

## Description
This PR establishes the modern development foundation for the Real-Time Forum project. It introduces Bun as the primary runtime and package manager for frontend tooling, configures Biome for unified linting and formatting, and sets up Vitest for frontend unit and integration testing.

Key changes:
- **Bun & Tooling Integration**: Installed Bun locally and configured it as the project's frontend runtime.
- **Robust Makefile**: Refined the `Makefile` to handle circular dependencies in `make deps` (bootstrapping Bun via npm) and improved `stop` safety by targeting exact binary names using `pkill -x` and returning to `go run` for development.
- **Biome Configuration**: Set up Biome 2.x with strict correctness rules. Configured ignores for legacy directories to ensure a clean linting baseline.
- **Vitest & Test Architecture**: Established a formal `Unit/Integration/E2E` testing structure in `AGENTS.md` and `SDS.md`. Initialized `web/tests/` and `web/SPA/` folder hierarchies with `.gitkeep` files.
- **Regression Coverage**: Migrated existing frontend integration tests to `web/tests/integration/` and verified they execute correctly via the updated `make test-frontend` command (which now runs both Go and JS tests).
- **CI/CD Pipeline**: Configured GitHub Actions (`ci.yml`) to enforce quality gates (lint, format, test) on every PR.

## Technical Decisions
- **Biome Pluralization**: Discovered that Biome 2.x CLI expects plural `includes` instead of `include` in some configuration blocks, which was addressed to achieve clean linting.
- **Legacy Exclusions**: Explicitly excluded `web/templates` and `web/errors` from Biome checks as they contain Go template syntax incompatible with standard HTML linting, ensuring the new SPA-focused tooling remains green.
- **Vitest Migration**: Adapted the previous VM-based custom test runner to Vitest by wrapping tests in `test()` blocks and using `beforeAll` for environment setup, preventing global namespace pollution.

## Verification Results
### Automated Tests
- `make test-backend`: **PASSED**
- `make test-frontend`: **PASSED** (4 tests: Image preview, Checkerboard for PNG, Lightbox expansion, Transparency presentation)
- `make lint`: **PASSED** (53 files checked)

### Manual Verification
- Verified `make deps` successfully bootstraps a fresh environment from an empty `node_modules`.
- Verified `make run` and `make stop` reliably manage `go run` processes without leaving orphaned bindings.
- Verified that `make test-frontend` correctly executes both Go integration tests and the new Vitest suite in `web/tests/`.

## Diff Summary (from main)
```text
 Makefile                                         |   31 +-
 AGENTS.md                                        |   14 +
 docs/SDS.md                                      |   10 +
 docs/ticket-tracker.md                           |    4 +-
 .github/workflows/ci.yml                         |   42 +
 package.json                                     |   23 +
 vitest.config.ts                                 |   10 +
 web/SPA/components/shared/.gitkeep               |    0
 web/SPA/core/api/.gitkeep                        |    0
 web/SPA/core/router/.gitkeep                     |    0
 web/SPA/features/auth/.gitkeep                   |    0
 web/SPA/features/feed/.gitkeep                   |    0
 web/tests/unit/.gitkeep                          |    0
 web/tests/integration/.gitkeep                   |    0
 web/tests/integration/frontend_behavior.test.mjs |  443 +++++
 web/tests/e2e/.gitkeep                           |    0
 ... (and other toolchain configuration files)
```
