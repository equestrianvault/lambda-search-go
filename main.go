package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type author struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

type link struct {
	Title string `json:"title"`
	Url   string `json:"url"`
}

type book struct {
	Id        int      `json:"id"`
	Title     string   `json:"title"`
	Edition   string   `json:"edition"`
	Image     string   `json:"img"`
	Rating    string   `json:"rating"`
	DateAdded string   `json:"dateAdded"`
	Expiry    string   `json:"expiry"`
	Authors   []author `json:"authors"`
	Links     []link   `json:"links"`
	Tags      []string `json:"tags"`
}

var (
	count  int64
	books  []book
	body   string
	logger *log.Logger
)

func main() {
	count = 0
	logger = log.New(os.Stderr, "http: ", log.LstdFlags)
	url := "https://raw.githubusercontent.com/equestrianvault/horsebooks-data/main/data/books.json"
	client := http.Client{
		Timeout: time.Second * 2,
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)

	if err != nil {
		log.Fatal(err)
	} else {
		logger.Println("No error constructing request.")
	}

	req.Header.Set("User-Agent", "go-server-project")

	res, getErr := client.Do(req)

	if getErr != nil {
		defer res.Body.Close()
	} else {
		logger.Println("No error executing Get request.")
	}

	if res.Body == nil {
		defer res.Body.Close()
	} else {
		logger.Println("No error examining body of Get request result.")
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		log.Fatal(readErr)
		logger.Println(readErr)
	} else {
		logger.Println("No error reading Get request body.")
	}

	// books = []book{}

	jsonErr := json.Unmarshal(body, &books)
	if jsonErr != nil {
		log.Fatal(jsonErr)
		logger.Println(readErr)
	} else {
		logger.Println("No error transforming Get request body to JSON.")
	}
	logger.Println("Books:", len(books))

	logger.Println("Starting service...")
	http.HandleFunc("/search", Search)
	http.ListenAndServe(":8080", nil)
}

func Search(w http.ResponseWriter, r *http.Request) {
	count++

	logger.Println("Request: ", count, " Num of books: ", len(books))

	w.Header().Set("X-Debug-Responses-Since-Boot", strconv.FormatInt(count, 10))
	w.Header().Set("X-Debug-Total-Num-Books", strconv.FormatInt(int64(len(books)), 10))

	// fmt.Fprintf(w, "Hello, %v!\n", r.URL.Path[1:])
	// for key, element := range r.URL.Query() {
	// 	fmt.Fprintf(w, "%v:[", key)
	// 	for order, value := range element {
	// 		fmt.Fprintf(w, "%v", value)
	// 		if order < len(element)-1 {
	// 			fmt.Fprintf(w, ",")
	// 		}
	// 	}
	// 	fmt.Fprintf(w, "]\n")
	// }

	bookObj, jsonMarshalErr := json.MarshalIndent(books, "", "  ")
	if jsonMarshalErr != nil {
		log.Fatal(jsonMarshalErr)
		logger.Println(jsonMarshalErr)
	} else {
		fmt.Fprintf(w, "%v", string(bookObj))
	}
}
