package main

import (
	"fmt"
	"time"
)

type Book struct {
	Title   string
	Self    *Link
	History []Checkout
}

type Checkout struct {
	Who    string
	Out    time.Time
	In     time.Time
	Review int
}

type Link struct {
	HRef string
	ID   string
}

var lib []Book

func init() {
	lib = append(lib, Book{
		Title: "Book 1",
		Self: &Link{
			HRef: "amazon.com",
			ID:   "1",
		},
	})
}

func get(id string) (book *Book) {
	for _, l := range lib {
		if l.Self.ID == id {
			book = &l
			break
		}
	}
	return
}

func getAll() []Book {
	return lib
}

func add(b Book) {
	lib = append(lib, b)
}

func checkout(b *Book, name string) error {
	h := len(b.History)
	if h != 0 {
		lastCheckout := b.History[h-1]
		if lastCheckout.In.IsZero() {
			return fmt.Errorf("this book is currently checked out to: [%s]", lastCheckout.Who)
		}
	}

	if len(name) == 0 {
		return fmt.Errorf("a name must be provided for checkout")
	}

	b.History = append(b.History, Checkout{
		Who: name,
		Out: time.Now(),
	})

	return nil
}

func checkin(b *Book, review int) error {
	h := len(b.History)
	if h == 0 {
		return fmt.Errorf("this book is not currently checked out")
	}

	if review < 1 || review > 5 {
		return fmt.Errorf("a review between 1 and 5 must be provided")
	}

	lastCheckout := b.History[h-1]
	if !lastCheckout.In.IsZero() {
		return fmt.Errorf("this book is not currently checked out")
	}

	b.History[h-1] = Checkout{
		Who:    lastCheckout.Who,
		Out:    lastCheckout.Out,
		In:     time.Now(),
		Review: review,
	}

	return nil
}
