# Makefile for NPB-GO MG Benchmark

# Go compiler
GO = go

# Source files
SOURCES = main.go mg_benchmark.go

# Default target
all: mg

# Build the MG benchmark (without tags - shows help message)
mg:
	$(GO) build -o mg $(SOURCES)

# Build with specific class
mg.S:
	$(GO) build -tags=S -o mg_S $(SOURCES)

mg.W:
	$(GO) build -tags=W -o mg_W $(SOURCES)

mg.A:
	$(GO) build -tags=A -o mg_A $(SOURCES)

mg.B:
	$(GO) build -tags=B -o mg_B $(SOURCES)

mg.C:
	$(GO) build -tags=C -o mg_C $(SOURCES)

mg.D:
	$(GO) build -tags=D -o mg_D $(SOURCES)

mg.E:
	$(GO) build -tags=E -o mg_E $(SOURCES)

# Build all classes
all-classes: mg.S mg.W mg.A mg.B mg.C mg.D mg.E

# Run benchmarks (no arguments needed - class is determined by build tags)
run.S: mg.S
	./mg_S

run.W: mg.W
	./mg_W

run.A: mg.A
	./mg_A

run.B: mg.B
	./mg_B

run.C: mg.C
	./mg_C

run.D: mg.D
	./mg_D

run.E: mg.E
	./mg_E

# Clean up
clean:
	rm -f mg mg_S mg_W mg_A mg_B mg_C mg_D mg_E

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

.PHONY: all mg mg.S mg.W mg.A mg.B mg.C mg.D mg.E all-classes run.S run.W run.A run.B run.C run.D run.E clean deps test fmt lint

