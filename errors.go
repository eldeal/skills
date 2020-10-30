package main

import (
	"fmt"
	"log"
	"net/http"
)

func readFailed(w http.ResponseWriter, err error) {
	log.Println(fmt.Sprintf("error reading request body: [%s]", err.Error))
	http.Error(w, "cannot read request body", http.StatusInternalServerError)
}

func bookNotFound(w http.ResponseWriter) {
	log.Println("book not found in list")
	http.Error(w, "book not found", http.StatusNotFound)
}

func unmarshalFailed(w http.ResponseWriter, err error) {
	log.Println(fmt.Sprintf("error returned from json unmarshal: [%s]", err.Error))
	http.Error(w, "cannot unmarshal json body", http.StatusInternalServerError)
}

func marshalFailed(w http.ResponseWriter, err error) {
	log.Println(fmt.Sprintf("error returned from json marshal: [%s]", err.Error))
	http.Error(w, "cannot marshal content to json", http.StatusInternalServerError)
}
