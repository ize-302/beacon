.PHONY: dev api sim test test-integration test-all

api:
	air -c .air.api.toml

sim:
	air -c .air.simulator.toml

dev:
	@trap 'kill 0' EXIT; \
	air -c .air.api.toml & \
	air -c .air.simulator.toml & \
	wait

# Unit tests. Fast, no dependencies, always run with the race detector since
# every bug this suite guards against is a concurrency bug.
test:
	go test -race ./...

# Black-box tests against a running stack (`docker compose up -d`). They skip
# rather than fail when nothing is listening.
test-integration:
	go test -race -tags=integration -count=1 ./test/...

test-all: test test-integration
