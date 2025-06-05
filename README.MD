# NPB-GO
NAS Parallel Benchmarks in Go

## Folders inside the project

NPB-SER - This directory contains the sequential version.


## Software requirements

go1.24.2 linux/amd64 or higher


## How to compile

Enter the directory from the version desired and execute:

```go
go build -o ./bin/ -tags=<CLASS>
```

<CLASS> are:

```
S: small for quick test purposes
W: workstation size (a 90's workstation; now likely too small)	
A, B, C: standard test problems; ~4X size increase going from one class to the next	
D, E: large test problems; ~16X size increase from each of the previous Classes  
```

The command above will compile all applications of the chosen version. To compile a specific application, it is possible to add the `--bin` flag:

```
RUSTFLAGS='--cfg class="_WORKLOAD"' cargo build --release --bin _BENCHMARK
```

_BENCHMARKs are:

```
ep
```
