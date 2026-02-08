# ===== CONFIG =====
CLASS ?= S
VERBOSE ?= 0

# List of valid kernels
# KERNELS := EP IS MG FT CG
KERNELS := EP IS MG FT CG

# Binary directory
BINDIR := bin

# Detect which target was called
KERNEL := $(firstword $(MAKECMDGOALS))

# ===== RULES =====

# Declare that kernels + other commands are always "phony"
.PHONY: $(KERNELS) clean build-all run

# Build template for each kernel
$(KERNELS):
	@if echo "$(KERNELS)" | grep -qw $@; then \
		if [ -d $@ ]; then \
			echo "==> Building kernel: $@ with CLASS=$(CLASS)"; \
			mkdir -p $(BINDIR); \
			if [ "$(VERBOSE)" -eq "1" ]; then \
				go build --tags=$(CLASS) -o $(BINDIR)/$@_$(CLASS) ./$@; \
			else \
				go build --tags=$(CLASS) -o $(BINDIR)/$@_$(CLASS) ./$@ >/dev/null 2>&1; \
			fi; \
			echo "==> Done! Executable: $(BINDIR)/$@_$(CLASS)"; \
		else \
			echo "ERROR: Kernel directory $@ not found!"; \
			exit 1; \
		fi \
	else \
		echo "ERROR: Invalid kernel '$@'. Valid options: $(KERNELS)"; \
		exit 1; \
	fi

# Build all kernels with the current CLASS
build-all:
	@for k in $(KERNELS); do \
		echo "==> Building $$k with CLASS=$(CLASS)"; \
		mkdir -p $(BINDIR); \
		if [ "$(VERBOSE)" -eq "1" ]; then \
			go build --tags=$(CLASS) -o $(BINDIR)/$$k_$(CLASS) ./$$k; \
		else \
			go build --tags=$(CLASS) -o $(BINDIR)/$$k_$(CLASS) ./$$k >/dev/null 2>&1; \
		fi; \
		echo "==> Done! Executable: $(BINDIR)/$$k_$(CLASS)"; \
	done

# Run a kernel (example: make run KERNEL=EP CLASS=A)
run:
	@if [ -z "$(KERNEL)" ]; then \
		echo "ERROR: Please specify KERNEL=<name> (ex.: make run KERNEL=EP CLASS=A)"; \
		exit 1; \
	elif echo "$(KERNELS)" | grep -qw $(KERNEL); then \
		EXE="$(BINDIR)/$(KERNEL)_$(CLASS)"; \
		if [ ! -f $$EXE ]; then \
			echo "==> Executable $$EXE not found! Building first..."; \
			mkdir -p $(BINDIR); \
			go build --tags=$(CLASS) -o $$EXE ./$(KERNEL); \
		fi; \
		echo "==> Running $$EXE"; \
		$$EXE; \
	else \
		echo "ERROR: Invalid kernel '$(KERNEL)'. Valid options: $(KERNELS)"; \
		exit 1; \
	fi

# Clean the bin folder
clean:
	@echo "==> Cleaning $(BINDIR)/"
	@rm -f $(BINDIR)/*
