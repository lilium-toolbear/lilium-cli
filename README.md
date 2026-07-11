# Lilium CLI

`gh`-style command line for ToolBear `/api/*` as yourself via OIDC.

## Install

```bash
# Nix
nix build
nix run . -- --help
nix profile install .

# Go
go install github.com/lilium-toolbear/lilium-cli/cmd/lilium@latest
# or from a checkout:
go build -o lilium ./cmd/lilium
```

Requires Go 1.20+ (or Nix).

## Configuration

| Variable / flag | Meaning |
|---|---|
| `LILIUM_HOST` / `--host` | ToolBear origin (default `https://lilium.kuma.homes`) |
| `LILIUM_CLIENT_ID` / `--client-id` | Override official public OIDC client UUID (default embedded) |
| `LILIUM_CONFIG_DIR` | Override credentials directory (default `~/.config/lilium`) |
| `LILIUM_CALLBACK_PORT` / `--callback-port` | Force loopback port (`3847`/`3848` tried by default) |

Credentials are stored at `~/.config/lilium/credentials.json` with mode `0600`. Refresh tokens are never logged.

## Auth

```bash
# optional staging overrides:
# export LILIUM_HOST="https://staging.example"
# export LILIUM_CLIENT_ID="<staging-uuid>"

lilium auth login            # desktop → browser loopback+PKCE; SSH → device code
lilium auth login --web      # force loopback
lilium auth login --device   # force device code
lilium auth status
lilium auth token            # print access token (agents)
lilium auth logout
```

Default scopes: `openid profile wallet:read` plus stock/market/turnip read+write.

Platform prerequisites (merged in dzmm_archive): public loopback redirects, secretless OIDC refresh for PUBLIC clients, device code grant, economy `require_scopes` on stock/market/turnip (+ wallet).

## Agent pattern

```bash
lilium auth login --device
lilium api GET /api/stock/portfolio
lilium wallet balance
```

`lilium api METHOD PATH` attaches Bearer auth and refreshes on expiry/401. For POST/PUT/PATCH, JSON body is read from stdin.

## Thin commands (V1)

| Command | API |
|---|---|
| `lilium wallet balance\|stats\|transactions` | `GET /api/wallet/...` |
| `lilium stock portfolio\|overview` | `GET /api/stock/...` |
| `lilium stock buy --symbol X --shares N` | `POST /api/stock/buy` |
| `lilium stock sell --symbol X --shares N` | `POST /api/stock/sell` |
| `lilium market listings\|my-orders` | `GET /api/market/...` |
| `lilium market buy\|sell\|cancel` | matching POSTs |
| `lilium turnip inventory\|farm\|buy\|sell` | `/api/turnip/...` |

Gift/poll/lottery/land/pal: use `lilium api` (no dedicated subcommands in V1).

## Official client seed

See `dzmm_archive` ops doc `docs/ops/lilium-cli-client.md`. Production client id is
embedded as `internal/config.DefaultClientID` (`589ed8c1-443f-4c13-ae4b-4891fb93de93`).
