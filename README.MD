# NPB-GO
NAS Parallel Benchmarks in Go

## Project structure

Each directory is independent and contains its own implemented version of the kernels

- `NPB-SER/` — Contains the sequential version of the NAS Parallel Benchmarks.
    - `common/` — Common utilities used by the benchmarks (random number generators, timers, result printing, etc.).
    - `EP/` — EP (Embarrassingly Parallel) benchmark.
    - Other kernels will be added progressively (`IS/`, `MG/`, etc.).
- `bin/` — Folder where compiled executables are stored.

## Software requirements

- Go **1.24.2** or higher, built for `linux/amd64`.




## How to build

The project provides a `Makefile` to simplify the build process.

### Building a specific benchmark

```bash
make <BENCHMARK> CLASS=<CLASS> 

```
Ex:
make EP CLASS=S

### Building a all benchmark

```bash
make <BENCHMARK> CLASS=<CLASS> 

```

Example:

make build-all CLASS=<CLASS>


### Runing a specific benchmark

```bash

make run KERNEL=<BENCHMARK> CLASS=<CLASS>

```


Eample:

make run KERNEL=EP CLASS=S

### Available Classes
```
S: small for quick test purposes
W: workstation size (a 90's workstation; now likely too small)
A, B, C: standard test problems; ~4X size increase going from one class to the next
D, E: large test problems; ~16X size increase from each of the previous Classes
```

Available Kernels are:

```
EP - Embarrassingly Parallel, floating-point operation capacity

```
