package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"
)

const (
	MAX_PAGE_SIZE            = 50
	DEFAULT_PAGE_SIZE        = 20
	DEFAULT_PAGE_NUM         = 1
	DEFAULT_QUERY_PARAM_NAME = "q"
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
	books  []book
	body   string
	logger *log.Logger
)

func main() {
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

func IsStringInBook(s string, b *book) (bool, error) {
	searchPattern, regexErr := regexp.Compile("(?i)" + regexp.QuoteMeta(s))
	if regexErr != nil {
		return true, regexErr
	}
	// don't filter if there are fewer than 3 characters
	if len(s) < 3 {
		return true, nil
	}

	// object checking -- probably need to throw an error
	if b == nil {
		return false, errors.New("Book Does Not Exist")
	}

	// tag search
	for _, tag := range b.Tags {
		if searchPattern.MatchString(tag) {
			return true, nil
		}
	}

	// title search
	if searchPattern.MatchString(b.Title) {
		return true, nil
	}

	// author search
	for _, author := range b.Authors {
		if searchPattern.MatchString(author.Name) {
			return true, nil
		}
	}

	// no match
	return false, nil
}

func Search(w http.ResponseWriter, r *http.Request) {
	logger.Println("Request: Num of books: ", len(books))
	w.Header().Set("X-Debug-Total-Num-Books", strconv.FormatInt(int64(len(books)), 10))

	filteredBooks := []book{}
	keys, ok := r.URL.Query()[DEFAULT_QUERY_PARAM_NAME]
	if !ok || len(keys) < 1 {
		filteredBooks = books
	} else {
		for _, buk := range books {
			check, err := IsStringInBook(keys[0], &buk)
			if err != nil {
				logger.Fatal(err)
			} else if check {
				filteredBooks = append(filteredBooks, buk)
			}
		}
		logger.Printf("Request: \"%v\"", keys[0])
		logger.Printf("Filtered Num of books: %v\n", len(filteredBooks))
	}

	bookObj, jsonMarshalErr := json.MarshalIndent(filteredBooks, "", "  ")

	if jsonMarshalErr != nil {
		log.Fatal(jsonMarshalErr)
		logger.Println(jsonMarshalErr)
	} else {
		fmt.Fprintf(w, "%v", string(bookObj))
	}
}
