This is a monorepo split into frontend and backend

Frontend is a vite app

Backend is an express app

There are no unit tests

## App structure (brief)
- `frontend/`: Vite + React UI dashboard to create/delete Podman containers and open VS Code Remote-SSH links per container.
- `backend/`: Express API + GitHub auth; `routes/containers.js` wraps `podman`, publishes SSH for new containers, and returns VS Code Remote-SSH URIs.
- Root: workspace config and shared docs (`package.json`, `.pnpm-workspace.yaml`, `README.md`).

## How the app works
- Auth: Express uses GitHub OAuth (or `NO_AUTH=true` for dev) and exposes `/api/me` for the UI session check.
- Session cookie: Express session cookie `maxAge` is 1 hour with `rolling: true` so TTL refreshes on each response.
- Session keepalive: The UI calls `/api/me` every 5 minutes to renew the cookie even when the tab is hidden.
- Dashboard: The React UI lists running app-owned containers and shows SSH info, password, and a VS Code Remote-SSH URI (including the repo path when available).
- Create container: `POST /api/containers` runs a devcontainers universal image, assigns a random animal name, sets the container hostname to that name, creates a random password, enables SSH, and (optionally) clones a GitHub repo into the container.
- SSH access: The container `sshd` listens on `2222` internally and Podman publishes that to a random host port. The API returns `user@host:port` and the password.
- Delete: `DELETE /api/containers/:id` removes the container via Podman.

## Learnings
- The devcontainers universal image runs `sshd` on port `2222`, not `22`. When publishing ports or generating VS Code Remote-SSH links, map and query `2222/tcp`.
- If SSH connects but aborts during banner exchange, check for mismatched internal SSH ports and ensure host key generation plus `sshd -D -e` in startup.
- Private repo cloning: GitHub OAuth requests `repo` scope and stores the access token server-side. `POST /api/containers` accepts `repoUrl`, clones via HTTPS using `http.extraheader` for auth, and adds repo labels at `podman run` time. The VS Code URI appends the repo path so the folder opens immediately on connect.