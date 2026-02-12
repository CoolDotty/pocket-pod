This app lets a developer run and manage dev environments in the cloud using containers.

run `npm run lint` when finalizing changes

# Project Structure

```text
.
|-- server/               Go app embedding PocketBase + frontend assets
|   |-- main.go           PocketBase bootstrap, middleware + route registration
|   |-- auth.go           Auth handlers (signup, login, logout, /me)
|   |-- cookies.go        HttpOnly cookie management + env-based config
|   |-- types.go          Payload structs, collection/role constants, sentinel errors
|   |-- routes.go         Cookie-to-header middleware + SPA static file serving
|   |-- assets.go         Asset loader contract
|   |-- assets_embed.go   Production asset embed via go:embed (web/dist)
|   |-- assets_dev.go     Dev asset loader for local frontend builds
|   |-- podman.go         Podman service + route registration for workspace APIs
|   |-- podman_actions.go Pod/container action implementations
|   |-- podman_parse.go   Podman CLI output parsing helpers
|   |-- podman_poller.go  Background poller for workspace/container status
|   |-- podman_workspace.go Workspace lifecycle and state helpers
|   |-- *_test.go         Unit tests for podman actions/poller/workspace
|   |-- migrations/
|   |   \-- 20260210_auth.go   Creates users (auth) + invites collections
|   \-- web/dist/         Frontend build output embedded in the Go binary
|-- web/                  React (Vite) frontend
|   |-- public/           Static public assets
|   |-- src/
|   |   |-- api/          PocketBase client + auth + podman queries/mutations/streaming
|   |   |-- components/   Shared UI + route guards
|   |   |-- context/      React context providers (Auth)
|   |   |-- hooks/        Reusable hooks (auth error handling)
|   |   |-- layouts/      App shell/layout components
|   |   |-- pages/        Route-level screens
|   |   |   |-- Dashboard/
|   |   |   |-- Login/
|   |   |   \-- Signup/
|   |   |-- types/        Shared TypeScript types
|   |   |-- App.tsx       App routes/composition
|   |   |-- main.tsx      React entry point
|   |   |-- index.css     Global styles
|   |   \-- vite-env.d.ts Vite type declarations
|   |-- index.html        Vite HTML entry
|   |-- eslint.config.js  Frontend ESLint config
|   |-- package.json      Frontend dependencies and scripts
|   |-- tsconfig.json     Frontend TypeScript config
|   |-- tsconfig.node.json Vite config TS support
|   \-- vite.config.ts    Vite config (proxy + build output)
|-- go.mod                Go module definition
|-- package.json          Root scripts for dev/build
|-- README.md             Project usage docs
\-- .gitignore            Repo ignore rules
```

# Go App Architecture

The server is a single Go binary powered by PocketBase. It embeds the frontend
build at compile time (`go:embed`) and serves it as an SPA with fallback routing.
In addition to auth routes, it exposes Podman-backed APIs for creating and
managing containerized dev workspaces.

## Auth Flow

- First signup auto-assigns the `admin` role; subsequent signups require an invite token.
- Auth tokens are issued via PocketBase's `NewAuthToken()` and stored in an HttpOnly
  cookie (`pb_auth`). A global middleware (`bindAuthCookieMiddleware`) copies the cookie
  into the `Authorization` header so PocketBase's built-in auth resolution populates
  `re.Auth` on every request.
- **Regression note**: `bindAuthCookieMiddleware` must run **before** PocketBase's
  `loadAuthToken` middleware. If its priority is too low, `re.Auth` won't populate on
  refresh even though the `pb_auth` cookie is present.
- Cookie security (Secure flag, TTL, domain) is configurable via environment variables:
  `AUTH_COOKIE_SECURE`, `AUTH_COOKIE_TTL_DAYS`, `AUTH_COOKIE_DOMAIN`.

## Key Conventions

- **Constants over magic strings**: Collection names (`CollectionUsers`, `CollectionInvites`)
  and roles (`RoleAdmin`, `RoleUser`) are defined in `types.go`.
- **Sentinel errors**: Domain errors (`errInviteRequired`, `errInviteUsed`, etc.) are
  defined as package-level vars and matched with `errors.Is()`.
- **Migrations are self-contained**: The `migrations` package uses inline string literals
  (not the main package constants) since migrations are point-in-time snapshots.

## Workspace + Tunnel Notes

- Default workspace image is Ubuntu: `docker.io/library/ubuntu:latest`.
- New workspace containers must mount user-scoped VS Code state:
  - Host: `./volumes/{userId}/.vscode`
  - Container: `/root/.vscode`
  - Use bind mount args via `--mount type=bind,...` with absolute host path.
- `./volumes/` must remain gitignored.
- Keepalive command for default workspace is:
  - `sh -lc "while true; do sleep 3600; done"`
- Tunnel bootstrap runs during workspace creation:
  - Install VS Code CLI only if `code` is not already installed.
  - Use direct download URLs:
    - x64: `https://code.visualstudio.com/sha/download?build=stable&os=cli-alpine-x64`
    - arm32: `https://code.visualstudio.com/sha/download?build=stable&os=cli-linux-armhf`
    - arm64: `https://code.visualstudio.com/sha/download?build=stable&os=cli-linux-arm64`
- Tunnel logging paths in container:
  - Bootstrap: `/tmp/pocketpod-vscode-bootstrap.log`
  - Tunnel runtime: `/tmp/pocketpod-vscode-tunnel.log`
- Tunnel auth URL is constant:
  - `https://github.com/login/device`
- Container tunnel state is exposed via `/podman/containers` and `/podman/containers/stream`:
  - `tunnelStatus`, `tunnelCode`, `tunnelMessage`
- Create workspace response includes tunnel snapshot:
  - `tunnel: { status, code?, message?, debug? }`
- Tunnel readiness rule:
  - Prefer runtime process check (`code tunnel` running) for `ready`.
  - Auth/device-code prompts from tunnel logs must surface as `blocked`.
  - Existing containers (created before current server run) must also be reconciled from logs.
