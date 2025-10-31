# CI/CD and Deployment Guide

This project ships with two GitHub Actions workflows that cover continuous integration and delivery:

- `.github/workflows/ci.yml` runs `go test`, vets the code, builds the binary, and performs a container build on every push or pull request targeting `main`.
- `.github/workflows/cd.yml` (1) rebuilds and pushes a production Docker image to GitHub Container Registry (GHCR) for every push to `main`, and (2) optionally deploys that image on a remote host over SSH.

Follow the steps below to enable the entire pipeline end-to-end.

## 1. Prepare the Container Registry

1. Ensure your GitHub account (or organisation) has GitHub Container Registry enabled.
2. Create a fine-grained personal access token (PAT) with **read & write** access to `packages` and **read** access to `contents`.
3. Decide which namespace you want to use in GHCR:
   - For user-scoped repositories: `ghcr.io/<your-gh-username>/poll-service`.
   - For organisations: `ghcr.io/<your-org>/poll-service`.

## 2. Configure Repository Secrets

Add the following secrets under **Settings > Secrets and variables > Actions > New repository secret**:

| Secret | Required | Description |
| --- | --- | --- |
| `REGISTRY_USERNAME` | Yes | GitHub username or organisation slug that owns the GHCR namespace. |
| `REGISTRY_PASSWORD` | Yes | PAT created in step 1 with `packages:read, packages:write`. |
| `REGISTRY_NAMESPACE` | Optional | Override for the namespace if it differs from `REGISTRY_USERNAME` (useful for org repos). |
| `PROD_ENV_FILE_BASE64` | Required for remote deploy | Base64-encoded contents of the production `.env` file. |
| `SSH_HOST` | Required for remote deploy | Hostname or IP address of the deployment server. |
| `SSH_USER` | Required for remote deploy | SSH user that can run Docker commands on the target server. |
| `SSH_PRIVATE_KEY` | Required for remote deploy | Private key (PEM) corresponding to an authorised key for `SSH_USER`. |
| `SSH_PORT` | Optional | Custom SSH port (defaults to `22` if omitted). |
| `REMOTE_WORKDIR` | Optional | Absolute path on the remote host where deployment assets should live (defaults to `/opt/poll-service`). |

> Note: To generate `PROD_ENV_FILE_BASE64`, run `base64 -w0 .env.production` (Linux/macOS) or `Get-Content .env.production | %{[Convert]::ToBase64String([Text.Encoding]::UTF8.GetBytes($_))}` (PowerShell) and paste the output as the secret value.

If any of the SSH or environment secrets are missing, the CD workflow will only build and push the container image and skip the remote deployment step.

## 3. Prepare the Remote Host

Ensure the target server meets the following requirements:

1. Docker Engine 20.10+ with the **Docker Compose v2 plugin** is installed (`docker compose version` should succeed).
2. The SSH user defined in `SSH_USER` can run Docker commands (add the user to the `docker` group or use root).
3. Outbound network access to `ghcr.io` is allowed so the server can pull images.
4. Ports `8080` (or your chosen service port) and `27017` (MongoDB) are open if you expect external traffic.

On the first run the workflow will copy `deploy/docker-compose.yaml` and the decoded `.env` file into the remote directory (default `/opt/poll-service`) and provision the stack using `docker compose`.

## 4. How Deployment Works

1. Every push to `main` triggers the `CD` workflow.
2. The workflow runs unit tests, builds the Docker image, and publishes two tags:
   - `ghcr.io/<namespace>/poll-service:latest`
   - `ghcr.io/<namespace>/poll-service:<git-sha>`
3. If the SSH and environment secrets are configured, the workflow:
   - Copies the Compose bundle and `.env` to the remote host.
   - Logs in to GHCR with the provided credentials.
   - Pulls the `:sha` image and brings up the stack with `docker compose`, wiring the container image in via the `POLL_IMAGE` environment variable.

Subsequent runs reuse the GHCR build cache and only update the services that changed.

## 5. Local Verification Tips

- Run `go test ./...` and `docker compose up --build` locally before pushing to reduce CI noise.
- If you need to simulate the remote deployment locally, export `POLL_IMAGE` to any valid image tag and run:
  ```bash
  POLL_IMAGE=ghcr.io/<namespace>/poll-service:<tag> docker compose -f deploy/docker-compose.yaml up -d
  ```

## 6. Extending the Pipeline

- **Additional registries:** Add another `docker/login-action` + `docker/build-push-action` block to `.github/workflows/cd.yml` using the desired registry credentials.
- **Staging environments:** Duplicate the deployment job with a different set of secrets (e.g., `STAGING_*`) and gate it on another branch.
- **Canary deploys / health checks:** Append commands to the final SSH step to run smoke tests or hit a health endpoint after `docker compose up`.

With the above configuration in place, simply push to `main` (or merge pull requests) to build, publish, and roll out the latest version of the Poll Service automatically.
