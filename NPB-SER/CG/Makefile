# Makefile for NPB-GO CG Benchmark

# Go compiler
GO = go

# Build flags
BUILD_FLAGS = -o cg

# Default target
all: cg

# Build the CG benchmark
cg:
	$(GO) build $(BUILD_FLAGS) main.go

# Build with specific class
cg.S: cg
	$(GO) build -tags S $(BUILD_FLAGS) main.go

cg.W: cg
	$(GO) build -tags W $(BUILD_FLAGS) main.go

cg.A: cg
	$(GO) build -tags A $(BUILD_FLAGS) main.go

cg.B: cg
	$(GO) build -tags B $(BUILD_FLAGS) main.go

cg.C: cg
	$(GO) build -tags C $(BUILD_FLAGS) main.go

cg.D: cg
	$(GO) build -tags D $(BUILD_FLAGS) main.go

cg.E: cg
	$(GO) build -tags E $(BUILD_FLAGS) main.go

# Run benchmarks
run: cg
	./cg

run.S: cg.S
	./cg S

run.W: cg.W
	./cg W

run.A: cg.A
	./cg A

run.B: cg.B
	./cg B

run.C: cg.C
	./cg C

run.D: cg.D
	./cg D

run.E: cg.E
	./cg E

# Clean up
clean:
	rm -f cg

# Install dependencies
deps:
	$(GO) mod tidy

# Test
test:
	$(GO) test ./...

# Format code
fmt:
	$(GO) fmt ./...

# Lint code
lint:
	$(GO) vet ./...

.PHONY: all cg run clean deps test fmt lint
