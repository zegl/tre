# tre

An experimental programming language backed by LLVM. The current goal is to make it compatible with Go.

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

## Features

### Types

Currently implemented:

* int (all signed types)
* string
* struct
* array

TODO:

* Slices
* Maps
* Bool

### Language features

* Functions
* * TODO: More than 1 return variable
* * TODO: Assign a function to a variable, functions as arguments etc.
* String features: `str[1]`, `str[1:5]`, `len(str)`
* Array features: `arr[1] = 123`, `arr[1]`, `len(arr)`
* Builtin methods
* * println (currently only works on strings)
* * printf (should be removed)
* * exit (shoould be removed)
* * len

TODO:

* Packages
* Interfaces
* Methods
* Pointers
