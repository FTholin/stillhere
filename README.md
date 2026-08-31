# stillhere

[![ci](https://github.com/FTholin/stillhere/actions/workflows/ci.yml/badge.svg)](https://github.com/FTholin/stillhere/actions/workflows/ci.yml)

> A small HTTP service written in Go, containerized and continuously deployed.

**Live:** https://stillhere-xxxx.onrender.com/healthz

---

## What it is

A minimal monitoring service: it exposes an HTTP API, tracks registered targets
and reports when one goes quiet. No external dependencies — everything runs on
the Go standard library.

## Endpoints

| Method | Path       | Response              |
| ------ | ---------- | --------------------- |
| `GET`  | `/healthz` | `{"status":"ok"}`     |
| `GET`  | `/version` | `{"version":"0.1.0"}` |

## Stack

- **Go 1.26** — standard library only (`net/http`, `log/slog`)
- **Structured logging** as JSON on stdout
- **Docker** — multi-stage build, distroless final image (~15 MB)
- **GitHub Actions** — formatting, vet, tests with the race detector, image build

## Running locally

```bash
go run ./cmd/stillhere
```

Then, in another terminal:

```bash
curl -i localhost:8080/healthz
```

The listening port is read from the `PORT` environment variable, defaulting to
`8080`:

```bash
PORT=3000 go run ./cmd/stillhere
```

## Tests

```bash
go test ./...             # all tests
go test -race ./...       # with the race detector
go vet ./...              # static analysis
gofmt -l .                # badly formatted files (empty output means all good)
```

## Docker

```bash
docker build -t stillhere .
docker run --rm -p 8080:8080 stillhere
```

With a custom port:

```bash
docker run --rm -p 3000:3000 -e PORT=3000 stillhere
```

## Layout

```
.
├── cmd/stillhere/       # entry point, route wiring
├── internal/api/        # HTTP handlers and their tests
├── Dockerfile           # multi-stage build to a distroless image
└── .github/workflows/   # continuous integration
```

## Deployment

Every push to `main` runs CI, then triggers a new deployment built from the
`Dockerfile`. No manual steps.

## License

MIT — see [LICENSE](LICENSE).