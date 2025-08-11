#!/usr/bin/make -f

########################################
### Simulations
### This Makefile runs simulation tests for the application in ./app.
### Prerequisites: Ensure 'runsim' is installed in $(BINDIR) (e.g., go install github.com/cosmos/tools/cmd/runsim@latest).
### Usage: make <target> [SIM_NUM_BLOCKS=<num>] [SIM_BLOCK_SIZE=<size>] [GENESIS_FILE=<path>]

BINDIR ?= $(or $(GOPATH),$(HOME)/go)/bin
SIMAPP = ./app
GENESIS_FILE ?= $(HOME)/.kiichain/config/genesis.json
TIMEOUT ?= 24h
SIM_NUM_BLOCKS ?= 500
SIM_BLOCK_SIZE ?= 200
SIM_COMMIT ?= true

# Ensure runsim is installed
runsim:
	@if ! command -v $(BINDIR)/runsim >/dev/null 2>&1; then \
		echo "Error: runsim not found in $(BINDIR). Install it with 'go install github.com/cosmos/tools/cmd/runsim@latest'"; \
		exit 1; \
	fi

test-sim-nondeterminism:
	@echo "Running non-determinism test..."
	@go test -mod=readonly $(SIMAPP) -run TestAppStateDeterminism -Enabled=true \
		-NumBlocks=100 -BlockSize=200 -Commit=true -Period=0 -v -timeout $(TIMEOUT)

test-sim-custom-genesis-fast:
	@echo "Running custom genesis simulation with $(GENESIS_FILE)..."
	@go test -mod=readonly $(SIMAPP) -run TestFullAppSimulation -Genesis=$(GENESIS_FILE) \
		-Enabled=true -NumBlocks=100 -BlockSize=200 -Commit=true -Seed=99 -Period=5 -v -timeout $(TIMEOUT)

test-sim-import-export: runsim
	@echo "Running application import/export simulation..."
	@$(BINDIR)/runsim -Jobs=4 -SimAppPkg=$(SIMAPP) -ExitOnFail 50 5 TestAppImportExport > import_export.log 2>&1

test-sim-after-import: runsim
	@echo "Running application simulation-after-import..."
	@$(BINDIR)/runsim -Jobs=4 -SimAppPkg=$(SIMAPP) -ExitOnFail 50 5 TestAppSimulationAfterImport > after_import.log 2>&1

test-sim-custom-genesis-multi-seed: runsim
	@echo "Running multi-seed custom genesis simulation with $(GENESIS_FILE)..."
	@$(BINDIR)/runsim -Genesis=$(GENESIS_FILE) -SimAppPkg=$(SIMAPP) -ExitOnFail 400 5 TestFullAppSimulation > custom_genesis_multi_seed.log 2>&1

test-sim-multi-seed-long: runsim
	@echo "Running long multi-seed application simulation..."
	@$(BINDIR)/runsim -Jobs=4 -SimAppPkg=$(SIMAPP) -ExitOnFail 500 50 TestFullAppSimulation > multi_seed_long.log 2>&1

test-sim-multi-seed-short: runsim
	@echo "Running short multi-seed application simulation..."
	@$(BINDIR)/runsim -Jobs=4 -SimAppPkg=$(SIMAPP) -ExitOnFail 50 10 TestFullAppSimulation > multi_seed_short.log 2>&1

test-sim-benchmark-invariants:
	@echo "Running simulation invariant benchmarks..."
	@go test -mod=readonly $(SIMAPP) -benchmem -bench=BenchmarkInvariants -run=^$ \
		-Enabled=true -NumBlocks=1000 -BlockSize=200 -Period=1 -Commit=true -Seed=57 -v -timeout $(TIMEOUT)

test-sim-benchmark:
	@echo "Running application benchmark for numBlocks=$(SIM_NUM_BLOCKS), blockSize=$(SIM_BLOCK_SIZE)..."
	@go test -mod=readonly -benchmem -run=^$$ $(SIMAPP) -bench ^BenchmarkFullAppSimulation$$ \
		-Enabled=true -NumBlocks=$(SIM_NUM_BLOCKS) -BlockSize=$(SIM_BLOCK_SIZE) -Commit=$(SIM_COMMIT) -timeout $(TIMEOUT)

test-sim-profile:
	@echo "Running application benchmark with profiling for numBlocks=$(SIM_NUM_BLOCKS), blockSize=$(SIM_BLOCK_SIZE)..."
	@go test -mod=readonly -benchmem -run=^$$ $(SIMAPP) -bench ^BenchmarkFullAppSimulation$$ \
		-Enabled=true -NumBlocks=$(SIM_NUM_BLOCKS) -BlockSize=$(SIM_BLOCK_SIZE) -Commit=$(SIM_COMMIT) -timeout $(TIMEOUT) -cpuprofile cpu.out -memprofile mem.out

clean:
	@echo "Cleaning up generated files..."
	@rm -f cpu.out mem.out *.log

.PHONY: runsim test-sim-nondeterminism test-sim-custom-genesis-fast test-sim-import-export \
        test-sim-after-import test-sim-custom-genesis-multi-seed test-sim-multi-seed-short \
        test-sim-multi-seed-long test-sim-benchmark-invariants test-sim-benchmark test-sim-profile clean
