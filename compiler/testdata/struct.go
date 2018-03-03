package main

type Book struct {
    book_id1 int
    book_id2 int
}

// main id1: 1111
// main id2: 2222
// info id1: 1111
// info id2: 2222
// info id1: 3333
// main id1: 3333
// info id1: 1111
// info id2: 2222
// info id1: 3333
// res id1: 3333
// res id2: 2222

func bookInfo(book Book) Book {
    printf("info id1: %d\n", book.book_id1)
    printf("info id2: %d\n", book.book_id2)

    book.book_id1 = 3333

    printf("info id1: %d\n", book.book_id1)

    return book
}

func main() {
    var bookiboy Book
    bookiboy.book_id1 = 1111
    bookiboy.book_id2 = 2222

    printf("main id1: %d\n", bookiboy.book_id1)
    printf("main id2: %d\n", bookiboy.book_id2)


    printf("main id1: %d\n", bookInfo(bookiboy).book_id1)
    res := bookInfo(bookiboy)

    printf("res id1: %d\n", res.book_id1)
    printf("res id2: %d\n", res.book_id2)
}