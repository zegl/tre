# tre

An experimental programming language backed by LLVM.

## Building 

```bash
# Build tre and run a test program
go build -i github.com/zegl/tre/cmd/tre && ./tre tests/tests/fib.tre && ./output-binary
```

## Example

Example program that calculates the fibonacci sequence.

```go
func fib(num i64) i64 {
    if num < 2 {
        return num
    }

    return fib(num-2) + fib(num-1)
}

func main() i64 {
    printf("%d\n", fib(34))
    return 0
}
```
