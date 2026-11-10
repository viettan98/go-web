package database

import (
	"encoding/csv"
	"encoding/json"
	"fmt"

	"os"
	"strconv"
	"strings"

	"web/internal/models"
)

func ImportBooksFromCSVFile(csvFilePath string) error {
	file, err := os.Open(csvFilePath)
	if err != nil {
		return fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to parse CSV data: %w", err)
	}

	if len(records) < 2 {
		return fmt.Errorf("CSV file contains no data rows")
	}

	headers := records[0]
	headerMap := make(map[string]int)
	for idx, name := range headers {
		headerMap[strings.TrimSpace(strings.ToLower(name))] = idx
	}

	bookMetaStore := NewDocumentStore("../data/books_nosql.json")
	importedCount := 0

	for _, row := range records[1:] {
		getValue := func(col string) string {
			if idx, exists := headerMap[col]; exists && idx < len(row) {
				return strings.TrimSpace(row[idx])
			}
			return ""
		}

		title := getValue("title")
		authorName := getValue("author")
		catName := getValue("category")
		desc := getValue("description")
		imageURL := getValue("image_url")

		if title == "" {
			continue
		}

		price, _ := strconv.ParseFloat(getValue("price"), 64)
		stock, _ := strconv.Atoi(getValue("stock"))

		var authorID string
		if authorName != "" {
			err := DB.QueryRow("SELECT id FROM authors WHERE name = ?", authorName).Scan(&authorID)
			if err != nil {
				var authorCount int
				DB.QueryRow("SELECT COUNT(*) FROM authors").Scan(&authorCount)
				authorID = fmt.Sprintf("a_%d", authorCount+1)
				_, _ = DB.Exec("INSERT INTO authors (id, name) VALUES (?, ?)", authorID, authorName)
			}
		}

		if catName == "" {
			catName = "Programming Books"
		}
		var categoryID string
		errCat := DB.QueryRow("SELECT id FROM categories WHERE name = ?", catName).Scan(&categoryID)
		if errCat != nil {
			var catCount int
			DB.QueryRow("SELECT COUNT(*) FROM categories").Scan(&catCount)
			categoryID = fmt.Sprintf("c_%d", catCount+1)
			slug := fmt.Sprintf("cat-%d", catCount+1)
			_, _ = DB.Exec("INSERT INTO categories (id, name, slug) VALUES (?, ?, ?)", categoryID, catName, slug)
		}

		var bookCount int
		DB.QueryRow("SELECT COUNT(*) FROM books").Scan(&bookCount)
		bookID := fmt.Sprintf("b_%d", bookCount+1)

		query := `
			INSERT INTO books (id, title, author_id, author, category_id, category, price_per_day, stock)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
			ON DUPLICATE KEY UPDATE title=title;
		`
		_, err = DB.Exec(query, bookID, title, authorID, authorName, categoryID, catName, price, stock)
		if err != nil {
			continue
		}

		docID := bookID
		metaBytes, _ := json.Marshal(struct {
			BookID      string `json:"book_id"`
			Title       string `json:"title"`
			ImageURL    string `json:"image_url"`
			Description string `json:"description"`
		}{
			BookID:      bookID,
			Title:       title,
			ImageURL:    imageURL,
			Description: desc,
		})

		bookMetaStore.SaveDocument(models.ContentDocument{
			ID:    docID,
			Title: "Meta for " + title,
			Body:  string(metaBytes),
		})

		importedCount++
	}

	fmt.Printf("Successfully imported %d books from CSV into MySQL & NoSQL!\n", importedCount)
	return nil
}