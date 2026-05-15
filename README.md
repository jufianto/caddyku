# caddyku

A CLI wrapper around [Caddy](https://caddyserver.com/) that manages a single shared reverse proxy for all your Docker projects on one VPS.

Caddy already handles everything — HTTPS, Let's Encrypt, hot reload. `caddyku` just makes it easy to register and remove domains across multiple projects without manually editing `Caddyfile` every time.

## How it works

```
~/projects/
├── caddy-proxy/          ← one Caddy instance for the whole VPS
│   ├── docker-compose.yml
│   └── Caddyfile         ← managed by caddyku
│
├── myapp/
│   ├── docker-compose.yml
│   └── caddyku.yaml      ← declares domains for this project
│
└── another-app/
    ├── docker-compose.yml
    └── caddyku.yaml
```

One `caddy-proxy` Docker Compose project runs Caddy on ports 80/443. Every app joins the shared `caddy-net` Docker network, and Caddy routes to each app by container name via Docker DNS.

`caddyku` edits the `Caddyfile` and reloads Caddy — that's it.

## Requirements

- Docker + Docker Compose
- A VPS or local machine running Linux, macOS, or Windows

## Installation

### Option 1 — Download binary (recommended, no Go required)

Go to the [Releases](https://github.com/jufianto/caddyku/releases/latest) page and download the binary for your platform, or use `curl` directly on your VPS:

**Linux (amd64)**
```bash
curl -sSL https://github.com/jufianto/caddyku/releases/latest/download/caddyku_linux_amd64.tar.gz \
  | tar -xz && sudo mv caddyku /usr/local/bin/
```

**Linux (arm64)** — e.g. Raspberry Pi, Oracle Cloud ARM
```bash
curl -sSL https://github.com/jufianto/caddyku/releases/latest/download/caddyku_linux_arm64.tar.gz \
  | tar -xz && sudo mv caddyku /usr/local/bin/
```

**macOS**
```bash
curl -sSL https://github.com/jufianto/caddyku/releases/latest/download/caddyku_darwin_arm64.tar.gz \
  | tar -xz && sudo mv caddyku /usr/local/bin/
```

Then verify:

```bash
caddyku --help
caddyku version
```

### Option 2 — Install with Go

If you have Go 1.21+ installed:

```bash
go install github.com/jufianto/caddyku@latest
```

Make sure `$GOPATH/bin` is in your `PATH`:

```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

## Shell Completion

`caddyku` uses Cobra's built-in completion support. You need to install the generated completion script once for your shell.

### zsh

```bash
mkdir -p ~/.zsh/completions
caddyku completion zsh > ~/.zsh/completions/_caddyku
```

Add this to `~/.zshrc` if it is not already configured:

```bash
fpath=(~/.zsh/completions $fpath)
autoload -Uz compinit
compinit
```

Then restart your shell, or run:

```bash
source ~/.zshrc
```

### bash

```bash
caddyku completion bash | sudo tee /etc/bash_completion.d/caddyku >/dev/null
```

Then restart your shell.

### fish

```bash
mkdir -p ~/.config/fish/completions
caddyku completion fish > ~/.config/fish/completions/caddyku.fish
```

Then restart your shell.

## Getting Started

### 1. Bootstrap the Caddy proxy (one time per VPS)

```bash
caddyku init
```

This creates `~/projects/caddy-proxy/` with a ready-to-use `docker-compose.yml` and an empty `Caddyfile`.

Then start Caddy:

```bash
cd ~/projects/caddy-proxy
docker compose up -d
```

Caddy is now running and ready to route traffic. HTTPS is automatic via Let's Encrypt for real domains. For local dev, use `.localhost` domains.

### 2. Set up a new app project

Have an existing app with a Docker Compose file? Run:

```bash
caddyku init-app \
  --dir ~/projects/myapp \
  --service backend \
  --container myapp-backend \
  --domain myapp.com \
  --upstream myapp-backend:8080
```

This does three things:
1. Patches your app's compose file to join `caddy-net` and sets `container_name` if needed
2. Creates `~/projects/myapp/caddyku.yaml` with your domain config
3. Adds the domain to the shared `Caddyfile` and reloads Caddy

Then restart your app so it joins the network:

```bash
cd ~/projects/myapp && docker compose up -d
```

### 3. Add a domain manually

```bash
caddyku add --domain myapp.com --upstream myapp-backend:8080
```

Or from a `caddyku.yaml` file:

```bash
caddyku add --config ~/projects/myapp/caddyku.yaml
```

### 4. Sync all projects at once

If you have many projects each with a `caddyku.yaml`, one command rebuilds the entire `Caddyfile`:

```bash
caddyku sync
```

This scans `~/projects/*/caddyku.yaml`, merges all domains into the `Caddyfile`, and reloads Caddy. Safe to run repeatedly — it only touches blocks it manages.

## Commands

| Command | Description |
|---|---|
| `caddyku init` | Bootstrap the caddy-proxy project |
| `caddyku add` | Add domain(s) to the Caddyfile |
| `caddyku remove` | Remove a domain from the Caddyfile |
| `caddyku sync` | Scan all projects and rebuild the Caddyfile |
| `caddyku reload` | Reload Caddy without downtime |
| `caddyku list` | List all domains in the current Caddyfile |
| `caddyku init-app` | Patch an app's docker-compose.yml and create caddyku.yaml |

### Global flags

| Flag | Default | Description |
|---|---|---|
| `--proxy-dir` | `~/projects/caddy-proxy` | Path to the caddy-proxy project |
| `--projects-dir` | `~/projects` | Root dir scanned by `sync` |
| `--caddy-service` | `caddy` | Service name in docker compose |

### `caddyku add`

```bash
# Single domain
caddyku add --domain app.com --upstream myapp-backend:8080

# Multiple domains from config file
caddyku add --config ~/projects/myapp/caddyku.yaml

# Add without reloading
caddyku add --domain app.com --upstream myapp-backend:8080 --no-reload
```

### `caddyku remove`

```bash
caddyku remove --domain app.com
```

### `caddyku init-app`

If you are already inside the app project, `--dir` is optional because it defaults to the current directory:

```bash
caddyku init-app --service backend --container myapp-backend
```

If your compose file is not named `docker-compose.yml`, pass `--compose-file`:

```bash
caddyku init-app --compose-file compose.yml --service backend
caddyku init-app --compose-file docker-compose.prod.yml --service backend
caddyku init-app --compose-file /srv/myapp/compose.production.yml --service backend
```

You can also put the compose/service/container/domain config in `caddyku.yaml` first:

```yaml
compose: docker-compose.prod.yml
service: backend
container: myapp-backend
domains:
  - domain: myapp.com
    upstream: myapp-backend:8080
```

Then run:

```bash
caddyku init-app
```

| Flag | Required | Description |
|---|---|---|
| `--dir` | no (default: `.`) | App project directory |
| `--compose-file` | no (default: `docker-compose.yml`) | Compose file to patch. Relative paths are resolved from `--dir`; absolute paths are used as-is. Can also be set with `compose:` in `caddyku.yaml`. |
| `--service` | yes, unless defined in `caddyku.yaml` | The service key inside the compose file to patch — e.g. if your compose has `services: backend:`, pass `--service backend`. Only this service will be added to `caddy-net`. |
| `--container` | no | The `container_name` to set only if the service does not already have one. Existing `container_name` is preserved by default. |
| `--domain` | no | Domain to register in the Caddyfile |
| `--upstream` | no | Where Caddy forwards requests, in `container_name:port` format. Usually the same as `--container` with the port your app listens on. |
| `--force-container` | no | Overwrite existing `container_name` if it differs from `--container` |

Example compose file before running `init-app`:

```yaml
services:
  backend:      # <-- this is --service backend
    image: myapp:latest
  db:
    image: postgres:16
```

After running `init-app --service backend --container myapp-backend`, the `backend` service joins `caddy-net`. If it does not already have `container_name`, it gets `container_name: myapp-backend`. If it already has a container name, caddyku keeps it unless you pass `--force-container`. The `db` service is untouched and stays internal.

### `caddyku sync`

```bash
caddyku sync
caddyku sync --projects-dir /srv/projects
```

## `caddyku.yaml` format

Each app project can have a `caddyku.yaml` to declare its domains:

```yaml
service: backend
container: myapp-backend
compose: docker-compose.prod.yml

domains:
  - domain: myapp.com
    upstream: myapp-backend:8080
  - domain: api.myapp.com
    upstream: myapp-backend:8080
  - domain: myapp.localhost
    upstream: myapp-backend:8080
```

The `compose` value is optional and defaults to `docker-compose.yml`. The `service` value is the key inside the compose file that should join `caddy-net`. The `container` value is optional; if your compose file already has `container_name`, caddyku keeps the existing value. The `upstream` value should point to the Docker DNS name Caddy can reach, usually `container_name:port`.

### Custom Caddy config

For simple apps, `upstream` is enough. For advanced Caddy behavior, use `body` instead. This lets you write custom Caddy directives while still letting `caddyku sync` manage the shared `Caddyfile`.

Example with a frontend route and a backend API route:

```yaml
domains:
  - domain: ujianku-stag.jufi.dev
    body: |
      # Proxy /uploads/* to the backend because uploaded files are stored there.
      handle /uploads/* {
          reverse_proxy ujianku-backend:8080
      }

      # Everything else goes to the frontend Nginx container.
      handle {
          reverse_proxy ujianku-frontend:80
      }

  - domain: ujianku-api-stag.jufi.dev
    body: |
      reverse_proxy ujianku-backend:8080 {
          header_up X-Real-IP {remote_host}
          header_up X-Forwarded-For {remote_host}
          header_up X-Forwarded-Proto {scheme}
      }
```

Generated Caddyfile:

```caddyfile
# BEGIN caddyku:ujianku
ujianku-stag.jufi.dev {
    # Proxy /uploads/* to the backend because uploaded files are stored there.
    handle /uploads/* {
        reverse_proxy ujianku-backend:8080
    }

    # Everything else goes to the frontend Nginx container.
    handle {
        reverse_proxy ujianku-frontend:80
    }
}

ujianku-api-stag.jufi.dev {
    reverse_proxy ujianku-backend:8080 {
        header_up X-Real-IP {remote_host}
        header_up X-Forwarded-For {remote_host}
        header_up X-Forwarded-Proto {scheme}
    }
}
# END caddyku:ujianku
```

Rules:

1. Use either `upstream` or `body` for a domain, not both.
2. `body` is raw Caddy config placed inside the domain block.
3. Caddy validates the final config during reload.

## Caddyfile format

`caddyku` uses marker comments to manage its blocks without touching anything else:

```
# BEGIN caddyku:myapp
myapp.com {
    reverse_proxy myapp-backend:8080
}

api.myapp.com {
    reverse_proxy myapp-backend:8080
}
# END caddyku:myapp

# BEGIN caddyku:another-app
another-app.com {
    reverse_proxy another-app-web:3000
}
# END caddyku:another-app
```

Anything outside these markers is left untouched — you can add custom Caddy config manually alongside `caddyku`-managed blocks.

## App docker-compose.yml requirements

For Caddy to route to your app, the service must:

1. Have a unique `container_name` across all projects
2. Join the `caddy-net` external network
3. **Not** expose ports to the host (Caddy handles that)

Example:

```yaml
services:
  backend:
    image: myapp:latest
    container_name: myapp-backend
    networks:
      - caddy-net
      - default

networks:
  caddy-net:
    external: true
    name: caddy-net
```

`caddyku init-app` writes this for you.

## TLS

Handled entirely by Caddy:

- **Real domain** (e.g. `myapp.com`) — Caddy automatically obtains a Let's Encrypt certificate. No config needed.
- **Local dev** (e.g. `myapp.localhost`) — Caddy serves over HTTP or with a self-signed cert automatically.

## License

MIT
