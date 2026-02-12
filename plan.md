# Workspace Create + Tunnel Auth State via Container Status Stream

## Summary
Implement tunnel bootstrap as part of workspace creation, then surface tunnel auth-block state through the same container status pipeline already used by `/podman/containers` and `/podman/containers/stream`.
If tunnel login is required, container data will include tunnel-block metadata (status + code), and frontend will show a global info block when at least one container is blocked.

## End-to-end flow
1. User creates workspace.
2. Backend creates and starts container.
3. Backend installs VS Code CLI in container (idempotent).
4. Backend starts `code tunnel` in container and returns `201 Created` immediately after bootstrap kickoff.
5. Post-boot monitor detects auth-required from tunnel logs.
6. Backend updates in-memory container state (same polling/streaming system) with tunnel auth-block info.
7. `/podman/containers` and websocket include that metadata.
8. Frontend detects any blocked container and renders info banner with login URL + code.

## Backend changes (`server/`)

### 1. Persistent VS Code token path
- Mount user-scoped host path into each workspace:
  - Host: `./volumes/{userId}/.vscode`
  - Container: `/root/.vscode`
- Create host dir before container create.
- Add `volumes/` to `.gitignore`.

### 2. Tunnel bootstrap during create
- In `createWorkspace(...)`, after container start:
  - `podman exec` install step (skip if `code` exists).
  - `podman exec` start tunnel step:
    - `code tunnel --accept-server-license-terms --name <workspaceNameOrId>`
    - Run detached and log to `/tmp/pocketpod-vscode-tunnel.log`.

### 3. Tunnel state model added to container API
- Extend `podmanContainer` with:
  - `tunnelStatus` (`ready | starting | blocked | failed`)
  - `tunnelCode` (optional)
  - `tunnelMessage` (optional)
- Keep existing fields unchanged.

### 4. Track tunnel state in existing service cache
- Add `podmanService.tunnelStateByContainerID map[string]...`.
- On poll/event merge, enrich container records with tunnel state before broadcasting/responding.
- On container remove event, delete tunnel state entry.

### 5. Post-boot tunnel monitor
- Background goroutine per created container:
  - Poll tunnel log and token file state for bounded retries.
  - If auth required, set:
    - `tunnelStatus=blocked`
    - `tunnelCode=<device code>`
    - `tunnelMessage="Authentication required"`
  - If token present/authenticated, set `tunnelStatus=ready` and clear code.
  - On unrecoverable failure, set `tunnelStatus=failed` with message.
- Trigger `schedulePoll(...)` or direct broadcast when tunnel state changes.

### 6. Create response contract
- `POST /podman/workspaces` returns `201` with existing workspace payload plus initial tunnel snapshot:
  - `tunnel: { status, code?, message? }`
- If auth becomes required shortly after, it will be reflected via container list + websocket updates (source of truth for evolving state).

## Frontend changes (`web/src/`)

### 1. Types + data handling
- Update container type and query/stream consumers to read `tunnelStatus`, `tunnelCode`, `tunnelMessage`.
- Update create response type with initial `tunnel` object.

### 2. Global auth info block
- On dashboard, if any container has `tunnelStatus === "blocked"`:
  - Show info panel:
    - fixed URL: `https://github.com/login/device`
    - code from one active blocked container (most recent/first)
    - container name for context
- If no blocked containers, hide panel.

### 3. Multi-container rule
- Show one active code in global block (not list all).

## Public API/interface/type additions
- `podmanContainer` JSON adds:
  - `tunnelStatus?: string`
  - `tunnelCode?: string`
  - `tunnelMessage?: string`
- Workspace create response adds:
  - `tunnel: { status: string; code?: string; message?: string }`

## Tests and scenarios

### 1. Backend unit tests
- Tunnel log parsing:
  - auth-required pattern extracts code
  - ready/authenticated detection
  - failed detection fallback
- Tunnel state merge into container snapshots.
- Event-driven cleanup of tunnel state on remove.
- Create flow:
  - always-start path
  - bootstrap command construction
  - response includes initial tunnel state

### 2. Frontend checks
- Dashboard renders auth info block only when blocked container exists.
- Uses fixed URL + active blocked code.
- Handles no-code/failed states gracefully.

### 3. Repo checks
- `go test ./...`
- `npm run lint`

## Assumptions/defaults locked
- Auth URL is constant: `https://github.com/login/device`.
- Tunnel process remains running after start.
- Token storage path is `./volumes/{userId}/.vscode`.
- `volumes/` is ignored by git.
