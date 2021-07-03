package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

type Book struct {
	Id                      string
	Series                  string
	SeriesNo                string
	Title                   string
	Author                  string
	AuthorLf                string
	AdditionalAuthors       string
	Isbn                    string
	Isbn13                  string
	MyRating                string
	AvgRating               string
	Publisher               string
	Binding                 string
	Pages                   string
	PublicationYear         string
	OriginalPublicationYear string
	ReadDate                string
	AddDate                 string
	Bookshelves             string
	State                   string
	ExclusiveShelf          string
	MyReview                string
	Spoiler string
	PrivateNotes string
	ReadCount string
	RecommendedFor string
	RecommendedBy string
}

func parseBookLine(record []string) Book {
	book := Book{}
	book.Id = record[0]
	title := record[1]
	if strings.Contains(title, "#") {
		// Handle series
		// e.g., `"Title": "Unsouled (Cradle, #1)"`

		splits := strings.Split(title, "(")
		book.Title = strings.TrimSpace(splits[0])

		if len(splits) > 1 {
			seriesSplit := strings.Split(splits[1], "#")
			series := strings.TrimRight(strings.TrimSpace(seriesSplit[0]), ",")
			book.Series = series
			if len(seriesSplit) > 1 {
				book.SeriesNo = strings.TrimRight(strings.TrimSpace(seriesSplit[1]), ")")
			}
		}

	} else {
		book.Title = record[1]
	}

	book.Author = record[2]
	book.AuthorLf = record[3]
	book.AdditionalAuthors = record[4]
	book.Isbn = strings.TrimLeft(strings.ReplaceAll(record[5], "\"", ""), "=")
	book.Isbn13 = strings.TrimLeft(strings.ReplaceAll(record[6], "\"", ""), "=")

	book.MyRating = record[7]
	if book.MyRating == "0" {
		book.MyRating = ""
	}

	book.AvgRating = record[8]
	book.Publisher = record[9]
	book.Binding = record[10]
	book.Pages = record[11]
	book.PublicationYear = record[12]
	book.OriginalPublicationYear = record[13]
	book.ReadDate = record[14]
	book.AddDate = record[15]
	book.Bookshelves = record[16]
	book.ExclusiveShelf = record[18]
	book.MyReview = record[19]
	book.Spoiler = record[20]
	book.PrivateNotes = record[21]
	book.ReadCount = record[22]
	book.RecommendedFor = record[23]
	book.RecommendedBy = record[24]

	switch book.ExclusiveShelf {
	case "read":
		book.State = "DONE"
	case "currently-reading":
		book.State = "INPROGRESS"
	case "to-read":
		book.State = "TODO"
	default:
		book.State = "TODO"

	}

	return book
}

func (v Book) String() string {
	var authors string
	if v.AdditionalAuthors != "" {
		authors = fmt.Sprintf("%s, %s", v.Author, v.AdditionalAuthors)
	} else {
		authors = v.Author
	}

	if len(v.Series) > 0 {
		return fmt.Sprintf("Title: %s (Serie: %s/%s) by %s on shelve: %s\n", v.Title, v.Series, v.SeriesNo, authors, v.Bookshelves)
	} else {
		return fmt.Sprintf("Title: %s by %s on shelve: %s\n", v.Title, authors, v.Bookshelves)
	}
}

func writeString(input string, key string) string {
	out := ""
	if len(input) > 0 {
		out = fmt.Sprintf(":%s: %s\n", key, input)
	}

	return out
}

func (b Book) ToOrgMode() string {
	var buffer strings.Builder

	// buffer.WriteString(fmt.Sprintf("** %s %s by %s\n", b.State, b.Title, b.Author))
	buffer.WriteString(fmt.Sprintf("** %s by %s\n", b.Title, b.Author))

	buffer.WriteString(":PROPERTIES:\n")

	buffer.WriteString(writeString(b.Title, "Title"))
	buffer.WriteString(writeString(b.Author, "Author"))
	buffer.WriteString(writeString(b.AdditionalAuthors, "AdditionalAuthors"))
	buffer.WriteString(writeString(b.Series, "Series"))
	buffer.WriteString(writeString(b.SeriesNo, "Series#"))
	buffer.WriteString(writeString(b.AvgRating, "AverageRating"))
	buffer.WriteString(writeString(b.MyRating, "MyRating"))
	buffer.WriteString(writeString(b.MyReview, "MyReview"))
	buffer.WriteString(writeString(b.PrivateNotes, "PrivateNotes"))
	buffer.WriteString(writeString(b.Publisher, "Publisher"))
	buffer.WriteString(writeString(b.OriginalPublicationYear, "FirstPublished"))
	buffer.WriteString(writeString(b.PublicationYear, "Published"))
	buffer.WriteString(writeString(b.Pages, "Pages"))
	buffer.WriteString(writeString(b.Bookshelves, "Bookshelves"))
	buffer.WriteString(writeString(b.ExclusiveShelf, "ExclusiveShelf"))
	buffer.WriteString(writeString(b.RecommendedBy, "RecommendedBy"))
	buffer.WriteString(writeString(b.RecommendedFor, "RecommendedFor"))
	buffer.WriteString(writeString(b.AddDate, "Added"))
	buffer.WriteString(writeString(b.ReadDate, "ReadDate"))
	buffer.WriteString(writeString(b.ReadCount, "ReadCount"))
	buffer.WriteString(writeString(b.Spoiler, "Spoiler"))
	buffer.WriteString(writeString(b.Isbn13, "ISBN13"))
	buffer.WriteString(writeString(b.Isbn, "ISBN"))
	buffer.WriteString(writeString(b.Id, "ID"))
	buffer.WriteString(writeString("folded", "visibility"))

	buffer.WriteString(":END:\n")

	return buffer.String()
}

func main() {

	log.Println("Starting importer")
	if len(os.Args) < 2 {
		fmt.Println("Please specify file to read")
		os.Exit(1)
	}

	file := os.Args[1]
	data, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Println("File reading error", err)
		os.Exit(1)
	}

	bookshelves := map[string][]Book{}

	//books := []Book{}
	r := csv.NewReader(bytes.NewReader(data))
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		book := parseBookLine(record)

		// Skip first line
		if book.Title != "Title" && book.Author != "Author" {
			bookshelves[book.ExclusiveShelf] = append(bookshelves[book.ExclusiveShelf], book)
		}

	}

	fmt.Println("#+TITLE: Books from Goodreads")
	fmt.Println("#+COMMENT: Imported by goodreads-to-org")
	fmt.Println("")

	shelves := 0
	noOfBooks := 0

	for shelf, books := range bookshelves {
		fmt.Println("* ", strings.ToUpper(shelf))
		shelves++
		noOfBooks += len(books)
		for _, book := range books {
			fmt.Println(book.ToOrgMode())
		}
		fmt.Println("\n")
	}
	log.Println("Parsing complete -", noOfBooks, "books on", shelves, "shelves")
}
