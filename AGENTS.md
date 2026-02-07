This is a monorepo split into frontend and backend

Frontend is a vite app

Backend is an express app

There are no unit tests

## App structure (brief)
- `frontend/`: Vite + React UI (dashboard to create/delete Podman containers).
- `backend/`: Express API + GitHub auth; `routes/containers.js` wraps `podman`.
- Root: workspace config and shared docs (`package.json`, `.pnpm-workspace.yaml`, `README.md`).
