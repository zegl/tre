package main

import "external"

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
	external.Printf("info id1: %d\n", book.book_id1)
	external.Printf("info id2: %d\n", book.book_id2)

	book.book_id1 = 3333

	external.Printf("info id1: %d\n", book.book_id1)

	return book
}

func main() {
	var bookiboy Book
	bookiboy.book_id1 = 1111
	bookiboy.book_id2 = 2222

	external.Printf("main id1: %d\n", bookiboy.book_id1)
	external.Printf("main id2: %d\n", bookiboy.book_id2)

	external.Printf("main id1: %d\n", bookInfo(bookiboy).book_id1)
	res := bookInfo(bookiboy)

	external.Printf("res id1: %d\n", res.book_id1)
	external.Printf("res id2: %d\n", res.book_id2)
}
