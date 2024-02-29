package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/JonecoBoy/multiCEP/external"
	"log"
	"net/http"
	"strings"
	"time"
)

func main() {

	mux := http.NewServeMux()
	// podia ter passado anonima
	mux.HandleFunc("/cep/", cepHandler)
	mux.HandleFunc("/", HomeHandler)

	log.Print("Listening...")
	http.ListenAndServe(":8080", mux)

}
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("pong"))
}

func cepHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.Split(r.URL.Path, "/")
	if len(path) < 3 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	cep := path[2]
	// remove separator if exists
	cep = strings.ReplaceAll(cep, "-", "")
	c, err := CepConcurrency(cep)
	if err != nil {
		fmt.Println(err.Error())
		w.Write([]byte(err.Error()))
		return
	}
	// Set the Content-Type header to application/json
	w.Header().Set("Content-Type", "application/json")

	jsonData, err := json.Marshal(c)
	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Print(string(jsonData))
	w.Write(jsonData)
}

func CepConcurrency(cep string) (external.Address, error) {
	c1 := make(chan external.Address)
	c2 := make(chan external.Address)

	go func() {
		//time.Sleep(time.Second * 2)
		data, err := external.BrasilApiCep(cep)
		if err != nil {
			log.Fatal(err)
		}
		c1 <- data
	}()
	go func() {
		//time.Sleep(time.Second * 2)
		data, err := external.ViaCep(cep)
		if err != nil {
			log.Fatal(err)
		}
		c2 <- data
	}()

	select {
	case msg := <-c1:
		return msg, nil
	case msg := <-c2:
		return msg, nil
	case <-time.After(time.Second * 1):
		return external.Address{}, errors.New("Timeout Reached, no API returned in time. CEP: " + cep)
	}
}
