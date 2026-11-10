package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
	"web/internal/database"
	"web/internal/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type AppHandler struct {
	CartStore     *database.DocumentStore
	BookMetaStore *database.DocumentStore
	ReviewStore   *database.DocumentStore
}

type Review struct {
	UserID    int    `json:"user_id"`
	Username  string `json:"username"`
	Rating    int    `json:"rating"`
	Comment   string `json:"comment"`
	CreatedAt string `json:"created_at"`
}

type CartItem struct {
	BookID   string  `json:"book_id"`
	Title    string  `json:"title"`
	Price    float64 `json:"price"`
	Quantity int     `json:"quantity"`
}

type BookMeta struct {
	BookID   string `json:"book_id"`
	Title    string `json:"title"`
	ImageURL string `json:"image_url"`
}

func NewAppHandler() *AppHandler {
	h := &AppHandler{
		CartStore:     database.NewDocumentStore("carts_nosql.json"),
		BookMetaStore: database.NewDocumentStore("books_nosql.json"),
		ReviewStore:   database.NewDocumentStore("reviews_nosql.json"),
	}
	h.syncAdminAccount()
	return h
}

func (h *AppHandler) syncAdminAccount() {
	if database.DB == nil {
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		log.Println("Failed to hash default admin password:", err)
		return
	}

	query := `
		INSERT INTO users (username, password, role) 
		VALUES ('admin', ?, 'admin')
		ON DUPLICATE KEY UPDATE role = 'admin';
	`
	_, err = database.DB.Exec(query, string(hashedPassword))
	if err != nil {
		log.Println("Error syncing admin account:", err)
	} else {
		log.Println("Admin account synchronized successfully!")
	}
}

func (h *AppHandler) RegisterRoutes(router *gin.Engine) {
	router.GET("/", h.RenderIndex)
	router.GET("/login", h.RenderLogin)
	router.GET("/dashboard", h.RenderUserDashboard)
	router.GET("/admin", h.RenderAdminDashboard)
	router.GET("/catalog", h.RenderCatalog)
	router.GET("/authors", h.RenderAuthors)
	router.GET("/book/:id", h.RenderBookDetail)
	router.GET("/profile", h.RenderProfile)
	router.GET("/checkout", h.RenderCheckout)

	api := router.Group("/api")
	{
		api.POST("/register", h.RegisterUser)
		api.POST("/login", h.LoginUser)
		api.GET("/books", h.GetBooks)
		api.POST("/books", h.CreateBook)
		api.GET("/books/:id", h.GetBookByID)

		api.POST("/reviews", h.AddBookReview)
		api.GET("/reviews/:book_id", h.GetBookReviews)
		api.POST("/cart", h.AddToCart)
		api.GET("/cart/:username", h.GetCart)
		api.POST("/profile/update", h.UpdateProfile)
		api.POST("/cart/remove", h.RemoveFromCart)
		api.POST("/checkout", h.ProcessCheckout)
		api.POST("/books/delete", h.DeleteBook)
		api.POST("/categories", h.CreateCategory)
		api.POST("/authors", h.CreateAuthor)
	}
}

func (h *AppHandler) RenderIndex(c *gin.Context)         { c.HTML(http.StatusOK, "index.html", nil) }
func (h *AppHandler) RenderLogin(c *gin.Context)         { c.HTML(http.StatusOK, "login.html", nil) }
func (h *AppHandler) RenderUserDashboard(c *gin.Context) { c.HTML(http.StatusOK, "user_dashboard.html", nil) }
func (h *AppHandler) RenderAdminDashboard(c *gin.Context){ c.HTML(http.StatusOK, "admin_dashboard.html", nil) }
func (h *AppHandler) RenderCatalog(c *gin.Context)       { c.HTML(http.StatusOK, "catalog.html", nil) }
func (h *AppHandler) RenderAuthors(c *gin.Context)       { c.HTML(http.StatusOK, "authors.html", nil) }
func (h *AppHandler) RenderBookDetail(c *gin.Context)    { c.HTML(http.StatusOK, "book_detail.html", nil) }
func (h *AppHandler) RenderProfile(c *gin.Context)       { c.HTML(http.StatusOK, "profile.html", nil) }
func (h *AppHandler) RenderCheckout(c *gin.Context)      { c.HTML(http.StatusOK, "checkout.html", nil) }

func (h *AppHandler) RegisterUser(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}
	_, err = database.DB.Exec("INSERT INTO users (username, password, role) VALUES (?, ?, 'user')", user.Username, string(hashedPassword))
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Registration successful"})
}

func (h *AppHandler) LoginUser(c *gin.Context) {
	var creds models.User
	if err := c.ShouldBindJSON(&creds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var storedHash, role string
	err := database.DB.QueryRow("SELECT password, role FROM users WHERE username = ?", creds.Username).Scan(&storedHash, &role)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Incorrect username or password"})
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(creds.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Incorrect username or password"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Login successful", "role": role, "username": creds.Username})
}

func (h *AppHandler) UpdateProfile(c *gin.Context) {
	var req struct {
		Username    string `json:"username"`
		NewPassword string `json:"new_password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.NewPassword != "" {
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
		_, err := database.DB.Exec("UPDATE users SET password = ? WHERE username = ?", string(hashedPassword), req.Username)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully"})
}

func (h *AppHandler) AddToCart(c *gin.Context) {
	var req struct {
		Username string  `json:"username"`
		BookID   string  `json:"book_id"`
		Title    string  `json:"title"`
		Price    float64 `json:"price"`
		Quantity int     `json:"quantity"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	docID := "cart_" + req.Username
	docs, _ := h.CartStore.GetAllDocuments()

	var items []CartItem
	for _, doc := range docs {
		if doc.ID == docID && doc.Body != "" {
			json.Unmarshal([]byte(doc.Body), &items)
			break
		}
	}

	qty := req.Quantity
	if qty <= 0 {
		qty = 1
	}

	found := false
	for i, item := range items {
		if item.BookID == req.BookID {
			items[i].Quantity += qty
			found = true
			break
		}
	}
	if !found {
		items = append(items, CartItem{BookID: req.BookID, Title: req.Title, Price: req.Price, Quantity: qty})
	}

	dataBytes, _ := json.Marshal(items)
	h.CartStore.SaveDocument(models.ContentDocument{
		ID:    docID,
		Title: "Cart for " + req.Username,
		Body:  string(dataBytes),
	})

	c.JSON(http.StatusOK, gin.H{"message": "Added to cart", "cart": items})
}

func (h *AppHandler) GetCart(c *gin.Context) {
	username := c.Param("username")
	docID := "cart_" + username

	docs, err := h.CartStore.GetAllDocuments()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"cart": []CartItem{}})
		return
	}

	var items []CartItem
	for _, doc := range docs {
		if doc.ID == docID && doc.Body != "" {
			json.Unmarshal([]byte(doc.Body), &items)
			break
		}
	}
	c.JSON(http.StatusOK, gin.H{"cart": items})
}

func (h *AppHandler) RemoveFromCart(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		BookID   string `json:"book_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	docID := "cart_" + req.Username
	docs, _ := h.CartStore.GetAllDocuments()

	var items []CartItem
	for _, doc := range docs {
		if doc.ID == docID && doc.Body != "" {
			json.Unmarshal([]byte(doc.Body), &items)
			break
		}
	}

	var updatedItems []CartItem
	for _, item := range items {
		if item.BookID != req.BookID {
			updatedItems = append(updatedItems, item)
		}
	}

	dataBytes, _ := json.Marshal(updatedItems)
	h.CartStore.SaveDocument(models.ContentDocument{
		ID:    docID,
		Title: "Cart for " + req.Username,
		Body:  string(dataBytes),
	})

	c.JSON(http.StatusOK, gin.H{"message": "Removed item from cart", "cart": updatedItems})
}

func (h *AppHandler) GetBooks(c *gin.Context) {
	rows, err := database.DB.Query("SELECT id, title, author, price_per_day, stock, category FROM books")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve book list"})
		return
	}
	defer rows.Close()

	metaDocs, _ := h.BookMetaStore.GetAllDocuments()
	metaMap := make(map[string]struct {
		ImageURL    string `json:"image_url"`
		Description string `json:"description"`
	})

	for _, doc := range metaDocs {
		var meta struct {
			BookID      string `json:"book_id"`
			ImageURL    string `json:"image_url"`
			Description string `json:"description"`
		}
		if json.Unmarshal([]byte(doc.Body), &meta) == nil {
			metaMap[meta.BookID] = struct {
				ImageURL    string `json:"image_url"`
				Description string `json:"description"`
			}{
				ImageURL:    meta.ImageURL,
				Description: meta.Description,
			}
		}
	}

	var books []map[string]interface{}
	for rows.Next() {
		var id, title, author, category string
		var price float64
		var stock int
		rows.Scan(&id, &title, &author, &price, &stock, &category)

		imgURL := "https://images.unsplash.com/photo-1543002588-bfa74002ed7e?auto=format&fit=crop&w=400&q=80"
		desc := ""
		if m, exists := metaMap[id]; exists {
			if m.ImageURL != "" {
				imgURL = m.ImageURL
			}
			desc = m.Description
		}

		if category == "" {
			category = "Programming Books"
		}

		books = append(books, gin.H{
			"id": id, "title": title, "author": author, "price": price, "stock": stock, "description": desc,
			"image_url": imgURL,
			"category":  category,
		})
	}
	c.JSON(http.StatusOK, gin.H{"books": books})
}

func (h *AppHandler) GetBookByID(c *gin.Context) {
	bookID := c.Param("id")
	var id, title, author, category string
	var price float64
	var stock int

	err := database.DB.QueryRow("SELECT id, title, author, price_per_day, stock, category FROM books WHERE id = ?", bookID).Scan(&id, &title, &author, &price, &stock, &category)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Book not found"})
		return
	}

	imgURL := "https://images.unsplash.com/photo-1543002588-bfa74002ed7e?auto=format&fit=crop&w=400&q=80"
	desc := ""
	metaDocs, _ := h.BookMetaStore.GetAllDocuments()
	for _, doc := range metaDocs {
		var meta struct {
			BookID      string `json:"book_id"`
			ImageURL    string `json:"image_url"`
			Description string `json:"description"`
		}
		if json.Unmarshal([]byte(doc.Body), &meta) == nil && meta.BookID == bookID {
			if meta.ImageURL != "" {
				imgURL = meta.ImageURL
			}
			desc = meta.Description
			break
		}
	}

	if category == "" {
		category = "Programming Books"
	}

	c.JSON(http.StatusOK, gin.H{"book": gin.H{
		"id": id, "title": title, "author": author, "price": price, "stock": stock, "description": desc,
		"image_url": imgURL, "category": category,
	}})
}

func (h *AppHandler) CreateBook(c *gin.Context) {
	var b struct {
		ID          string  `json:"id"`
		Title       string  `json:"title"`
		Author      string  `json:"author"`
		Category    string  `json:"category"`
		Price       float64 `json:"price"`
		Stock       int     `json:"stock"`
		Description string  `json:"description"`
		ImageURL    string  `json:"image_url"`
	}
	if err := c.ShouldBindJSON(&b); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	bookID := b.ID
	if bookID == "" || bookID == "undefined" || bookID[:2] != "b_" {
		var count int
		database.DB.QueryRow("SELECT COUNT(*) FROM books").Scan(&count)
		bookID = fmt.Sprintf("b_%d", count+1)
	}

	var authorID string
	if b.Author != "" {
		err := database.DB.QueryRow("SELECT id FROM authors WHERE name = ?", b.Author).Scan(&authorID)
		if err != nil {
			var authorCount int
			database.DB.QueryRow("SELECT COUNT(*) FROM authors").Scan(&authorCount)
			authorID = fmt.Sprintf("a_%d", authorCount+1)
			_, _ = database.DB.Exec("INSERT INTO authors (id, name) VALUES (?, ?)", authorID, b.Author)
		}
	}

	categoryName := b.Category
	if categoryName == "" {
		categoryName = "Programming Books"
	}
	var categoryID string
	errCat := database.DB.QueryRow("SELECT id FROM categories WHERE name = ?", categoryName).Scan(&categoryID)
	if errCat != nil {
		var catCount int
		database.DB.QueryRow("SELECT COUNT(*) FROM categories").Scan(&catCount)
		categoryID = fmt.Sprintf("c_%d", catCount+1)
		slug := fmt.Sprintf("cat-%d", catCount+1)
		_, _ = database.DB.Exec("INSERT INTO categories (id, name, slug) VALUES (?, ?, ?)", categoryID, categoryName, slug)
	}

	query := `
		INSERT INTO books (id, title, author_id, author, category_id, category, price_per_day, stock)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE title=?, author_id=?, author=?, category_id=?, category=?, price_per_day=?, stock=?;
	`
	_, err := database.DB.Exec(query, 
		bookID, b.Title, authorID, b.Author, categoryID, categoryName, b.Price, b.Stock,
		b.Title, authorID, b.Author, categoryID, categoryName, b.Price, b.Stock,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save book to database: " + err.Error()})
		return
	}

	docID := "book_meta_" + bookID
	metaBytes, _ := json.Marshal(struct {
		BookID      string `json:"book_id"`
		Title       string `json:"title"`
		ImageURL    string `json:"image_url"`
		Description string `json:"description"`
	}{
		BookID:      bookID,
		Title:       b.Title,
		ImageURL:    b.ImageURL,
		Description: b.Description,
	})

	h.BookMetaStore.SaveDocument(models.ContentDocument{
		ID:    docID,
		Title: "Meta for " + b.Title,
		Body:  string(metaBytes),
	})

	c.JSON(http.StatusCreated, gin.H{"message": "Book saved & NoSQL synchronized successfully", "book_id": bookID})
}

func (h *AppHandler) DeleteBook(c *gin.Context) {
	var req struct {
		ID string `json:"id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := database.DB.Exec("DELETE FROM books WHERE id = ?", req.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete book from database: " + err.Error()})
		return
	}

	docID := "book_meta_" + req.ID
	_ = h.BookMetaStore.SaveDocument(models.ContentDocument{
		ID:    docID,
		Title: "Deleted Meta",
		Body:  "{}",
	})

	c.JSON(http.StatusOK, gin.H{"message": "Book deleted successfully"})
}

func (h *AppHandler) GetBookReviews(c *gin.Context) {
	bookID := c.Param("book_id")
	docs, _ := h.ReviewStore.GetAllDocuments()
	var reviews []Review
	for _, doc := range docs {
		if doc.ID == "review_"+bookID && doc.Body != "" {
			json.Unmarshal([]byte(doc.Body), &reviews)
			break
		}
	}
	c.JSON(http.StatusOK, gin.H{"reviews": reviews})
}

func (h *AppHandler) AddBookReview(c *gin.Context) {
	var req struct {
		BookID   string `json:"book_id"`
		Username string `json:"username"`
		Rating   int    `json:"rating"`
		Comment  string `json:"comment"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	username := req.Username
	if username == "" {
		username = "Anonymous"
	}

	var userID int
	err := database.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
	if err != nil {
		userID = 0
	}

	targetID := "review_" + req.BookID
	docs, _ := h.ReviewStore.GetAllDocuments()
	var reviews []Review
	for _, doc := range docs {
		if doc.ID == targetID && doc.Body != "" {
			json.Unmarshal([]byte(doc.Body), &reviews)
			break
		}
	}

	newReview := Review{
		UserID:    userID,
		Username:  username,
		Rating:    req.Rating,
		Comment:   req.Comment,
		CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
	}
	reviews = append(reviews, newReview)

	dataBytes, _ := json.Marshal(reviews)
	
	h.ReviewStore.SaveDocument(models.ContentDocument{
		ID:    targetID,
		Title: "Reviews for " + req.BookID,
		Body:  string(dataBytes),
	})

	c.JSON(http.StatusCreated, gin.H{"reviews": reviews})
}

func (h *AppHandler) CreateCategory(c *gin.Context) {
	var req struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category name"})
		return
	}

	var catCount int
	database.DB.QueryRow("SELECT COUNT(*) FROM categories").Scan(&catCount)
	catID := fmt.Sprintf("c_%d", catCount+1)
	slug := fmt.Sprintf("cat-%d", catCount+1)

	_, err := database.DB.Exec("INSERT INTO categories (id, name, slug) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE name=name", catID, req.Name, slug)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save category: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Category saved successfully", "id": catID})
}

func (h *AppHandler) CreateAuthor(c *gin.Context) {
	var req struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid author name"})
		return
	}

	var authorCount int
	database.DB.QueryRow("SELECT COUNT(*) FROM authors").Scan(&authorCount)
	authorID := fmt.Sprintf("a_%d", authorCount+1)

	_, err := database.DB.Exec("INSERT INTO authors (id, name) VALUES (?, ?) ON DUPLICATE KEY UPDATE name=name", authorID, req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save author: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Author saved successfully", "id": authorID})
}

func (h *AppHandler) ProcessCheckout(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Fullname string `json:"fullname"`
		Address  string `json:"address"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	docID := "cart_" + req.Username
	docs, _ := h.CartStore.GetAllDocuments()
	var items []CartItem
	for _, doc := range docs {
		if doc.ID == docID && doc.Body != "" {
			json.Unmarshal([]byte(doc.Body), &items)
			break
		}
	}

	if len(items) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cart is empty!"})
		return
	}

	var total float64
	for _, item := range items {
		total += item.Price * float64(item.Quantity)
	}

	tx, err := database.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database system error"})
		return
	}

	res, err := tx.Exec("INSERT INTO orders (username, fullname, address, total_amount) VALUES (?, ?, ?, ?)", req.Username, req.Fullname, req.Address, total)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
		return
	}

	orderID, _ := res.LastInsertId()

	for _, item := range items {
		_, err = tx.Exec("INSERT INTO order_items (order_id, book_id, quantity, price) VALUES (?, ?, ?, ?)", orderID, item.BookID, item.Quantity, item.Price)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save order details"})
			return
		}
	}

	tx.Commit()

	h.CartStore.SaveDocument(models.ContentDocument{
		ID:    docID,
		Title: "Cart for " + req.Username,
		Body:  "[]",
	})

	c.JSON(http.StatusOK, gin.H{"message": "Checkout successful!", "order_id": orderID})
}