#!/bin/bash

# Build script for NPB-GO IS parallel benchmark with channels
# Usage: ./build.sh [CLASS]
# CLASS can be S, A, B, C, D (default: S)

CLASS=${1:-S}

echo "Building NPB-GO IS parallel benchmark (with channels) for class $CLASS..."

# Check if class is supported
case $CLASS in
    S|A|B|C|D)
        echo "Class $CLASS is supported"
        ;;
    *)
        echo "Error: Class $CLASS is not supported. Available classes: S, A, B, C, D"
        exit 1
        ;;
esac

# Clean previous build
rm -f is_parallel

# Build with the specified class
go build -tags $CLASS -o is_parallel main.go

if [ $? -eq 0 ]; then
    echo "Build successful! Executable: is_parallel"
    echo "Run with: ./is_parallel"
    echo ""
    echo "Class $CLASS specifications:"
    case $CLASS in
        S) echo "  Size: 65,536 keys" ;;
        A) echo "  Size: 8,388,608 keys" ;;
        B) echo "  Size: 33,554,432 keys" ;;
        C) echo "  Size: 134,217,728 keys" ;;
        D) echo "  Size: 2,147,483,648 keys" ;;
    esac
    echo "  Processors: $(nproc) (limited to 8 max)"
else
    echo "Build failed!"
    exit 1
fi
