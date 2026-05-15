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

Go to the [Releases](https://github.com/jufi/caddyku/releases/latest) page and download the binary for your platform, or use `curl` directly on your VPS:

**Linux (amd64)**
```bash
curl -sSL https://github.com/jufi/caddyku/releases/latest/download/caddyku_linux_amd64.tar.gz \
  | tar -xz && sudo mv caddyku /usr/local/bin/
```

**Linux (arm64)** — e.g. Raspberry Pi, Oracle Cloud ARM
```bash
curl -sSL https://github.com/jufi/caddyku/releases/latest/download/caddyku_linux_arm64.tar.gz \
  | tar -xz && sudo mv caddyku /usr/local/bin/
```

**macOS**
```bash
curl -sSL https://github.com/jufi/caddyku/releases/latest/download/caddyku_darwin_arm64.tar.gz \
  | tar -xz && sudo mv caddyku /usr/local/bin/
```

Then verify:

```bash
caddyku --help
```

### Option 2 — Install with Go

If you have Go 1.21+ installed:

```bash
go install github.com/jufi/caddyku@latest
```

Make sure `$GOPATH/bin` is in your `PATH`:

```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

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

Have an existing app with a `docker-compose.yml`? Run:

```bash
caddyku init-app \
  --dir ~/projects/myapp \
  --service backend \
  --container myapp-backend \
  --domain myapp.com \
  --upstream myapp-backend:8080
```

This does three things:
1. Patches your app's `docker-compose.yml` to join `caddy-net` and sets `container_name`
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

```bash
caddyku init-app \
  --dir ~/projects/myapp \
  --service backend \
  --container myapp-backend \
  --domain myapp.com \
  --upstream myapp-backend:8080
```

| Flag | Required | Description |
|---|---|---|
| `--dir` | no (default: `.`) | App project directory |
| `--service` | yes | Service name in docker-compose.yml |
| `--container` | yes | Container name (used by Caddy for DNS routing) |
| `--domain` | no | Domain to register |
| `--upstream` | no | Upstream in `container:port` format |

### `caddyku sync`

```bash
caddyku sync
caddyku sync --projects-dir /srv/projects
```

## `caddyku.yaml` format

Each app project can have a `caddyku.yaml` to declare its domains:

```yaml
domains:
  - domain: myapp.com
    upstream: myapp-backend:8080
  - domain: api.myapp.com
    upstream: myapp-backend:8080
  - domain: myapp.localhost
    upstream: myapp-backend:8080
```

The `upstream` value must match the `container_name` of the service in your `docker-compose.yml`.

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
