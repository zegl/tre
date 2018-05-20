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

- [x] int (all signed types)
- [x] string
- [x] struct
- [x] array
- [x] slice
- [ ] map
- [x] bool
- [ ] func

### Language features

- [ ] Functions
- - [x] Named functions
- - [x] Methods
- - [ ] More than 1 return variable
- - [ ] First class functions: Assign a function to a variable, functions as arguments etc.
- [x] Strings
- [x] Arrays
- [x] Builtin functions
- - [x] println (to be removed)
- - [x] printf (to be removed)
- - [x] exit (to be removed)
- - [x] len (string, array, slice) TODO: Maps
- - [x] cap (slice)
- - [x] append (slice)
- [ ] Packages
- [x] Methods
- [ ] Pointers
- [ ] Interfaces
- - [x] Empty interfaces
- - [x] Values receiver methods
- - [ ] Pointer receiver methods

## Dependencies

* [clang](https://clang.llvm.org/) - (only tested with version 6)
* [llir/llvm](https://github.com/llir/llvm)
