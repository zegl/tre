# tre

> A LLVM backed Go compiler.

`tre` is built in Go and can compile a subset of Go code to LLVM IR. Clang is
used to compile the IR to an executable.

## Building 

```bash
# Build tre and run a test program
go build -i github.com/zegl/tre/cmd/tre && ./tre tests/tests/fib.tre && ./output-binary
```

## Example

Example program that calculates the fibonacci sequence.

```go
func fib(num int) int {
    if num < 2 {
        return num
    }

    return fib(num-2) + fib(num-1)
}

func main() {
    printf("%d\n", fib(34))
}
```

More examples of what's possible can be found in the [compiler testdata](https://github.com/zegl/tre/tree/master/compiler/testdata).

## Features

### Types

- [x] int
- [x] string
- [x] struct
- [x] array
- [x] slice
- [ ] [map](https://github.com/zegl/tre/issues/34)
- [x] bool
- [x] func
- [ ] [chan](https://github.com/zegl/tre/issues/78)

### Language features

- [ ] [first class func](https://github.com/zegl/tre/issues/36) 
- [ ] packages
- [x] methods
- [x] pointers
- [x] interfaces
- [ ] [chan](https://github.com/zegl/tre/issues/78)
- [ ] [goroutines](https://github.com/zegl/tre/issues/77)
- [x] if/else if/else
- [ ] switch

## Dependencies

* [clang](https://clang.llvm.org/) - Supports 8.0 and 9.0
* [llir/llvm](https://github.com/llir/llvm)
