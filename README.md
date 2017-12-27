# tre

An experimental programming language backed by LLVM.

## Building 

```bash
# Build tre and run a test program
go build -i && ./tre tests/printf.tre > tmp.ll && clang tmp.ll && ./a.out
```
