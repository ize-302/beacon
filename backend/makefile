.PHONY: dev api sim

api:
	air -c .air.api.toml

sim:
	air -c .air.simulator.toml

dev:
	@trap 'kill 0' EXIT; \
	air -c .air.api.toml & \
	air -c .air.simulator.toml & \
	wait
