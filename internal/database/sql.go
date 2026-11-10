package database

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

var DB *sql.DB

func InitMySQL(dataSourceName string) {
	var err error
	DB, err = sql.Open("mysql", dataSourceName)
	if err != nil {
		log.Fatalf("Failed to open MySQL connection: %v", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatalf("MySQL database is not reachable: %v", err)
	}

	createTables()
	seedDefaultData()
	log.Println("MySQL database initialized successfully.")
}

func createTables() {
	userTableQuery := `
	CREATE TABLE IF NOT EXISTS users (
		id INT AUTO_INCREMENT PRIMARY KEY,
		username VARCHAR(50) UNIQUE NOT NULL,
		password VARCHAR(255) NOT NULL,
		role VARCHAR(20) DEFAULT 'user',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		INDEX idx_username (username)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`
	_, err := DB.Exec(userTableQuery)
	if err != nil {
		log.Fatalf("Failed to create users table: %v", err)
	}

	catTableQuery := `
	CREATE TABLE IF NOT EXISTS categories (
		id VARCHAR(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci PRIMARY KEY,
		name VARCHAR(100) NOT NULL UNIQUE,
		slug VARCHAR(100) NOT NULL UNIQUE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`
	_, err = DB.Exec(catTableQuery)
	if err != nil {
		log.Fatalf("Failed to create categories table: %v", err)
	}

	authorTableQuery := `
	CREATE TABLE IF NOT EXISTS authors (
		id VARCHAR(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci PRIMARY KEY,
		name VARCHAR(150) NOT NULL UNIQUE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		INDEX idx_author_name (name)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`
	_, err = DB.Exec(authorTableQuery)
	if err != nil {
		log.Fatalf("Failed to create authors table: %v", err)
	}

	bookTableQuery := `
	CREATE TABLE IF NOT EXISTS books (
		id VARCHAR(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci PRIMARY KEY,
		title VARCHAR(255) NOT NULL,
		author_id VARCHAR(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL,
		author VARCHAR(255) NOT NULL,
		category_id VARCHAR(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL,
		category VARCHAR(100) DEFAULT 'Programming Books',
		price_per_day DECIMAL(10, 2) NOT NULL,
		stock INT DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE SET NULL,
		FOREIGN KEY (author_id) REFERENCES authors(id) ON DELETE SET NULL,
		INDEX idx_book_title (title),
		INDEX idx_book_category (category)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`
	_, err = DB.Exec(bookTableQuery)
	if err != nil {
		log.Fatalf("Failed to create books table: %v", err)
	}

	orderTableQuery := `
	CREATE TABLE IF NOT EXISTS orders (
		id INT AUTO_INCREMENT PRIMARY KEY,
		username VARCHAR(50) NOT NULL,
		fullname VARCHAR(100) NOT NULL,
		address TEXT NOT NULL,
		total_amount DECIMAL(10, 2) NOT NULL,
		status VARCHAR(50) DEFAULT 'Completed',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		INDEX idx_order_user (username)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`
	_, err = DB.Exec(orderTableQuery)

	orderItemsTableQuery := `
	CREATE TABLE IF NOT EXISTS order_items (
		id INT AUTO_INCREMENT PRIMARY KEY,
		order_id INT NOT NULL,
		book_id VARCHAR(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
		quantity INT NOT NULL,
		price DECIMAL(10, 2) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE,
		FOREIGN KEY (book_id) REFERENCES books(id) ON DELETE CASCADE
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`
	_, err = DB.Exec(orderItemsTableQuery)
}

func seedDefaultData() {
	hashedPass, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err == nil {
		_, _ = DB.Exec(`
			INSERT INTO users (username, password, role) 
			VALUES ('admin', ?, 'admin')
			ON DUPLICATE KEY UPDATE role = 'admin';
		`, string(hashedPass))
	}

	
}