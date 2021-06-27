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
	// case insensitive REGEX matcher! Looks for the provided string as a partial search
	searchPattern, regexErr := regexp.Compile("(?i)" + regexp.QuoteMeta(s))
	// pass all results if there is a regex error
	if regexErr != nil {
		log.Fatal(regexErr)
		return true, regexErr
	}

	// don't filter if there are fewer than 3 characters
	if len(s) < 3 {
		return true, nil
	}

	// do we have a valid book?
	if b == nil {
		return false, errors.New("Book Does Not Exist")
	}

	// tag search, include when matched
	for _, tag := range b.Tags {
		if searchPattern.MatchString(tag) {
			return true, nil
		}
	}

	// title search, include when matched
	if searchPattern.MatchString(b.Title) {
		return true, nil
	}

	// author search, include when matched
	for _, author := range b.Authors {
		if searchPattern.MatchString(author.Name) {
			return true, nil
		}
	}

	// no match - don't include
	return false, nil
}

func Search(w http.ResponseWriter, r *http.Request) {
	logger.Println("Request: Num of books: ", len(books))
	w.Header().Set("X-Debug-Total-Num-Books", strconv.FormatInt(int64(len(books)), 10))

	filteredBooks := []book{}

	// find the query string
	keys, ok := r.URL.Query()[DEFAULT_QUERY_PARAM_NAME]

	if !ok || len(keys) < 1 {
		filteredBooks = books
	} else {

		// Only return results that match all of the keys
		for _, buk := range books {
			metaCheck, _ := IsStringInBook(keys[0], &buk)
			for _, key := range keys {
				check, err := IsStringInBook(key, &buk)
				if err != nil {
					log.Fatal(err)
				}
				metaCheck = (check && metaCheck)
			}

			if metaCheck {
				filteredBooks = append(filteredBooks, buk)
			}
		}

		// debug output
		logger.Printf("Request: \"%v\"", keys)
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
