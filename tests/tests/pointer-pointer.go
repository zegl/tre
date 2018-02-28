package main

func main() {
    i := 100
    iptr1 := &i
    iptr2 := &iptr1
    iptr3 := &iptr2

    // 100
    // 100
    // 100
    printf("%d\n", i)
    printf("%d\n", **iptr2)
    printf("%d\n", ***iptr3)
}
