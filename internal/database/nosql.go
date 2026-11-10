package database

import (
	"encoding/json"
	"path/filepath"
	"os"
	"sync"
	"time"
	"web/internal/models"
)

type DocumentStore struct {
	mu       sync.Mutex
	filePath string
}

type BookNoSQL struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	ImageURL    string `json:"image_url"`
	Description string `json:"description"`
}

type ReviewNoSQL struct {
	UserID    int    `json:"user_id"`
	Username  string `json:"username"`
	Rating    int    `json:"rating"`
	Comment   string `json:"comment"`
	CreatedAt string `json:"created_at"`
}

type CartItemNoSQL struct {
	BookID   string  `json:"book_id"`
	Title    string  `json:"title"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
}

type TempCartNoSQL struct {
	CartID    string          `json:"cart_id"`
	UserID    string          `json:"user_id"`
	Items     []CartItemNoSQL `json:"items"`
	CreatedAt string          `json:"created_at"`
}

type AdminAuditLog struct {
	LogID         string `json:"log_id"`
	AdminUsername string `json:"admin_username"`
	ActionType    string `json:"action_type"`
	Details       string `json:"details"`
	Timestamp     string `json:"timestamp"`
}

func NewDocumentStore(fileName string) *DocumentStore {
	dataDir := "data"
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		os.MkdirAll(dataDir, 0755)
	}
	
	filePath := filepath.Join(dataDir, fileName)
	return &DocumentStore{filePath: filePath}
}

func (ds *DocumentStore) SaveDocument(doc models.ContentDocument) error {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	docs, _ := ds.GetAllDocuments()
	found := false
	for i, d := range docs {
		if d.ID == doc.ID {
			docs[i] = doc
			found = true
			break
		}
	}
	if !found {
		docs = append(docs, doc)
	}

	fileData, err := json.MarshalIndent(docs, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(ds.filePath, fileData, 0644)
}

func (ds *DocumentStore) GetAllDocuments() ([]models.ContentDocument, error) {
	if _, err := os.Stat(ds.filePath); os.IsNotExist(err) {
		return []models.ContentDocument{}, nil
	}

	fileData, err := os.ReadFile(ds.filePath)
	if err != nil {
		return nil, err
	}

	var docs []models.ContentDocument
	if len(fileData) == 0 {
		return []models.ContentDocument{}, nil
	}

	err = json.Unmarshal(fileData, &docs)
	return docs, err
}

func SaveBookNoSQL(book BookNoSQL) error {
	store := NewDocumentStore("books_nosql.json")
	dataBytes, err := json.Marshal(book)
	if err != nil {
		return err
	}
	return store.SaveDocument(models.ContentDocument{
		ID:    book.ID,
		Title: book.Title,
		Body:  string(dataBytes),
	})
}

func SaveReviewNoSQL(bookID string, review ReviewNoSQL) error {
	store := NewDocumentStore("reviews_nosql.json")
	if review.CreatedAt == "" {
		review.CreatedAt = time.Now().Format("2006-01-02 15:04:05")
	}

	docs, _ := store.GetAllDocuments()
	var reviews []ReviewNoSQL
	for _, doc := range docs {
		if doc.ID == "review_"+bookID && doc.Body != "" {
			json.Unmarshal([]byte(doc.Body), &reviews)
			break
		}
	}
	reviews = append(reviews, review)
	dataBytes, _ := json.Marshal(reviews)

	return store.SaveDocument(models.ContentDocument{
		ID:    "review_" + bookID,
		Title: "Reviews for " + bookID,
		Body:  string(dataBytes),
	})
}

func SaveCartNoSQL(cart TempCartNoSQL) error {
	store := NewDocumentStore("carts_nosql.json")
	if cart.CreatedAt == "" {
		cart.CreatedAt = time.Now().Format("2006-01-02 15:04:05")
	}
	dataBytes, err := json.Marshal(cart)
	if err != nil {
		return err
	}
	return store.SaveDocument(models.ContentDocument{
		ID:    cart.CartID,
		Title: "Cart for " + cart.UserID,
		Body:  string(dataBytes),
	})
}

func LogAdminAction(adminUsername, actionType, details string) {
	logStore := NewDocumentStore("admin_audit_log.json")
	docs, _ := logStore.GetAllDocuments()
	var logs []AdminAuditLog
	for _, doc := range docs {
		if doc.ID == "audit_logs" && doc.Body != "" {
			json.Unmarshal([]byte(doc.Body), &logs)
			break
		}
	}

	logID := "log_" + time.Now().Format("20060102150405")
	logs = append(logs, AdminAuditLog{
		LogID:         logID,
		AdminUsername: adminUsername,
		ActionType:    actionType,
		Details:       details,
		Timestamp:     time.Now().Format("2006-01-02 15:04:05"),
	})

	dataBytes, _ := json.Marshal(logs)
	logStore.SaveDocument(models.ContentDocument{
		ID:    "audit_logs",
		Title: "Admin Activity Audit Log",
		Body:  string(dataBytes),
	})
}