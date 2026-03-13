# M0: Project Skeleton

## Goal

Bootable project with CI, linting, and directory structure. No functionality yet.

## Tasks

- [ ] `go mod init github.com/mdjarv/db`
- [ ] Create directory structure per [architecture.md](architecture.md)
- [ ] Add root cobra command (`cmd/root.go`) with version flag
- [ ] Add `main.go` that calls `cmd.Execute()`
- [ ] Configure `golangci-lint` (`.golangci.yml`) — enable `govet`, `errcheck`, `staticcheck`, `revive`
- [ ] Add `Makefile` with targets: `build`, `lint`, `test`, `test-integration`
- [ ] Add GitHub Actions CI: lint + test on push/PR
- [ ] Add `.gitignore` (Go binary, IDE files, `.env`)
- [ ] Add basic `CLAUDE.md` with project-specific build/test/lint commands

## Acceptance Criteria

- `go build ./...` succeeds
- `./db --version` prints version
- `make lint` passes
- `make test` passes (no tests yet, but exits 0)
- CI runs on push

## Dependencies

None — this is the starting point.

## Estimated Effort

Small. Single developer, a few hours.
