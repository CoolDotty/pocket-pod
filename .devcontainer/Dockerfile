FROM mcr.microsoft.com/devcontainers/typescript-node:latest

# Some devcontainer setups attempt to chown to a "user" account; ensure it exists.
RUN if ! id -u user >/dev/null 2>&1; then useradd -m -s /bin/bash user; fi