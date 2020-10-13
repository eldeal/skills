package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func main() {

	r := mux.NewRouter()

	r.HandleFunc("/library", createBook).Methods("POST")
	r.HandleFunc("/library", listBooks).Methods("GET")
	r.HandleFunc("/library/{id}", getBook).Methods("GET")

	r.HandleFunc("/library/{id}/checkout", checkoutBook).Methods("POST")
	r.HandleFunc("/library/{id}/checkout", checkinBook).Methods("PUT")

	http.ListenAndServe(":8080", r)
}

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
func createBook(w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		log.Println("error reading request body to create book")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Unmarshal to the shape of Response struct
	var book Book
	err = json.Unmarshal(b, &book)
	if err != nil {
		log.Println("error returned from json unmarshal when creating book")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	lib = append(lib, book)

	w.Header().Set("content-type", "application/json")
	w.Write(b)
}

func listBooks(w http.ResponseWriter, r *http.Request) {
	b, err := json.Marshal(lib)
	if err != nil {
		log.Println("error returned from json marshal")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(b)

}

func getBook(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	var book *Book
	for _, l := range lib {
		if l.Self.ID == id {
			book = &l
			break
		}
	}

	if book == nil {
		log.Println("book not found in list")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	b, err := json.Marshal(book)
	if err != nil {
		log.Println("error returned from json marshal")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

func checkoutBook(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	return
}

func checkinBook(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	return
}
