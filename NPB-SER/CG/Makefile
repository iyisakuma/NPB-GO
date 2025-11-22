# Makefile for NPB-GO CG Benchmark

# Go compiler
GO = go

# Source files
SOURCES = main.go cg_benchmark.go

# Default target
all: cg

# Build the CG benchmark (without tags - shows help message)
cg:
	$(GO) build -o cg $(SOURCES)

# Build with specific class
cg.S:
	$(GO) build -tags=S -o cg_S $(SOURCES)

cg.W:
	$(GO) build -tags=W -o cg_W $(SOURCES)

cg.A:
	$(GO) build -tags=A -o cg_A $(SOURCES)

cg.B:
	$(GO) build -tags=B -o cg_B $(SOURCES)

cg.C:
	$(GO) build -tags=C -o cg_C $(SOURCES)

cg.D:
	$(GO) build -tags=D -o cg_D $(SOURCES)

cg.E:
	$(GO) build -tags=E -o cg_E $(SOURCES)

# Build all classes
all-classes: cg.S cg.W cg.A cg.B cg.C cg.D cg.E

# Run benchmarks (no arguments needed - class is determined by build tags)
run.S: cg.S
	./cg_S

run.W: cg.W
	./cg_W

run.A: cg.A
	./cg_A

run.B: cg.B
	./cg_B

run.C: cg.C
	./cg_C

run.D: cg.D
	./cg_D

run.E: cg.E
	./cg_E

# Clean up
clean:
	rm -f cg cg_S cg_W cg_A cg_B cg_C cg_D cg_E

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

.PHONY: all cg cg.S cg.W cg.A cg.B cg.C cg.D cg.E all-classes run.S run.W run.A run.B run.C run.D run.E clean deps test fmt lint
