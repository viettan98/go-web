package database

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

func InitMySQL(dataSourceName string) error {
	var err error
	DB, err = sql.Open("mysql", dataSourceName)
	if err != nil {
		return err
	}

	if err = DB.Ping(); err != nil {
		return err
	}

	log.Println("Connected to MySQL successfully!")
	
	if err = runMigrations(); err != nil {
		return err
	}

	return nil
}

func runMigrations() error {
	queries := []string{
		// 1. Authors table
		`CREATE TABLE IF NOT EXISTS authors (
			id VARCHAR(50) PRIMARY KEY,
			name VARCHAR(255) NOT NULL UNIQUE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`,

		// 2. Categories table
		`CREATE TABLE IF NOT EXISTS categories (
			id VARCHAR(50) PRIMARY KEY,
			name VARCHAR(255) NOT NULL UNIQUE,
			slug VARCHAR(255) NOT NULL
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`,

		// 3. Books table
		`CREATE TABLE IF NOT EXISTS books (
			id VARCHAR(50) PRIMARY KEY,
			title VARCHAR(255) NOT NULL,
			author_id VARCHAR(50),
			author VARCHAR(255),
			category_id VARCHAR(50),
			category VARCHAR(255),
			price_per_day DECIMAL(10,2) NOT NULL,
			stock INT NOT NULL DEFAULT 0,
			FOREIGN KEY (author_id) REFERENCES authors(id) ON DELETE SET NULL,
			FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE SET NULL
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`,

		// 4. Users table
		`CREATE TABLE IF NOT EXISTS users (
			id INT AUTO_INCREMENT PRIMARY KEY,
			username VARCHAR(100) NOT NULL UNIQUE,
			password VARCHAR(255) NOT NULL,
			role VARCHAR(50) NOT NULL DEFAULT 'user'
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`,

		// 5. Orders table
		`CREATE TABLE IF NOT EXISTS orders (
			id INT AUTO_INCREMENT PRIMARY KEY,
			username VARCHAR(100) NOT NULL,
			fullname VARCHAR(255) NOT NULL,
			address TEXT NOT NULL,
			total_amount DECIMAL(10,2) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`,

		// 6. Order items table
		`CREATE TABLE IF NOT EXISTS order_items (
			id INT AUTO_INCREMENT PRIMARY KEY,
			order_id INT NOT NULL,
			book_id VARCHAR(50) NOT NULL,
			quantity INT NOT NULL,
			price DECIMAL(10,2) NOT NULL,
			FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE,
			FOREIGN KEY (book_id) REFERENCES books(id) ON DELETE CASCADE
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`,
	}

	for _, q := range queries {
		_, err := DB.Exec(q)
		if err != nil {
			log.Printf("Migration error: %v", err)
			return err
		}
	}

	// Safely add indexes if they do not exist
	indexQueries := []string{
		`SET @exist := (SELECT COUNT(*) FROM information_schema.statistics WHERE table_schema = DATABASE() AND table_name = 'books' AND index_name = 'idx_book_title');`,
		`SET @sql := IF(@exist = 0, 'ALTER TABLE books ADD INDEX idx_book_title (title)', 'SELECT 1');`,
		`PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;`,

		`SET @exist := (SELECT COUNT(*) FROM information_schema.statistics WHERE table_schema = DATABASE() AND table_name = 'books' AND index_name = 'idx_book_category');`,
		`SET @sql := IF(@exist = 0, 'ALTER TABLE books ADD INDEX idx_book_category (category_id)', 'SELECT 1');`,
		`PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;`,

		`SET @exist := (SELECT COUNT(*) FROM information_schema.statistics WHERE table_schema = DATABASE() AND table_name = 'orders' AND index_name = 'idx_order_user');`,
		`SET @sql := IF(@exist = 0, 'ALTER TABLE orders ADD INDEX idx_order_user (username)', 'SELECT 1');`,
		`PREPARE stmt FROM @sql; EXECUTE stmt; DEALLOCATE PREPARE stmt;`,
	}

	for _, q := range indexQueries {
		_, _ = DB.Exec(q)
	}

	log.Println("Database migrations and indexes applied successfully!")
	return nil
}