package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// sql queries
const (
	GetAllUsers          = `SELECT id, first_name, last_name FROM users`
	GetAllAvailableBooks = `SELECT id, title, quantity FROM books WHERE quantity > 0`
	AddUser              = `INSERT INTO users (first_name, last_name) VALUES ($1, $2) RETURNING id`
	AddBook              = `INSERT INTO books (title, quantity) VALUES ($1, $2) RETURNING id`
	BorrowBook           = `INSERT INTO borrowings (user_id, book_id) VALUES ($1, $2)`
	GetBookQuantity      = `SELECT quantity FROM books WHERE id = $1`
	DecreaseBookQty      = `UPDATE books SET quantity = quantity - 1 WHERE id = $1`
	ReturnBook           = `DELETE FROM borrowings WHERE user_id = $1 AND book_id = $2 LIMIT 1`
	IncreaseBookQty      = `UPDATE books SET quantity = quantity + 1 WHERE id = $1`
	CreateUsersDB        = `CREATE TABLE IF NOT EXISTS users (id SERIAL PRIMARY KEY, first_name VARCHAR(50), last_name VARCHAR(50))`
	CreateBooksDB        = `CREATE TABLE IF NOT EXISTS books ( id SERIAL PRIMARY KEY, title VARCHAR(100), quantity INT)`
	CreateBorrowDB       = `CREATE TABLE IF NOT EXISTS borrowings (id SERIAL PRIMARY KEY, user_id INT REFERENCES users(id), book_id INT REFERENCES books(id), borrow_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP);`

	AddBooksIfNotExist = `WITH new_books AS (VALUES ('The Great Gatsby', 5), ('1984', 3), ('To Kill a Mockingbird', 7), ('The Catcher in the Rye', 0), ('Pride and Prejudice', 10))
						  INSERT INTO books (title, quantity) SELECT column1 AS title, column2 AS quantity FROM new_books WHERE NOT EXISTS (SELECT 1 FROM books WHERE books.title = new_books.column1);`
)

// User db formats
type User struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type Book struct {
	ID           int    `json:"id"`
	BookName     string `json:"book_name"`
	BookQuantity int    `json:"book_quantity"`
}

var BorrowData struct {
	UserID int `json:"user_id"`
	BookID int `json:"book_id"`
}

var db *sql.DB

// run server function
func main() {
	var err error = godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	db = dbConnect()

	initDB()

	route := mux.NewRouter()

	route.HandleFunc("/displayUsers", displayUser).Methods("GET")
	route.HandleFunc("/displayBooks", displayAwaliableBooks).Methods("GET")
	route.HandleFunc("/addUser", addUser).Methods("POST")
	route.HandleFunc("/borrowBook", borrowBook).Methods("POST")
	route.HandleFunc("/returnBook", returnBook).Methods("POST")

	defer db.Close()

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", route))
}

// db init
// creates DB if it doesn't exist
func initDB() {
	queries := []string{CreateUsersDB, CreateBooksDB, CreateBorrowDB}
	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			log.Fatal(err)
		}
	}
	db.Exec(AddBooksIfNotExist)
}

// connects to DB using credentials from .env file
func dbConnect() (db *sql.DB) {
	var err error

	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	dbUrl := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPassword, dbHost, dbPort, dbName)
	db, err = sql.Open("postgres", dbUrl)

	if err != nil {
		log.Fatal(err)
	}

	return db
}

// display functions
// displays all users as id, first name, last name
func displayUser(writer http.ResponseWriter, request *http.Request) {
	rows, err := db.Query(GetAllUsers)

	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}

	defer rows.Close()

	var usrs []User
	for rows.Next() {
		var usr User
		if err := rows.Scan(&usr.ID, &usr.FirstName, &usr.LastName); err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}
		usrs = append(usrs, usr)
	}

	json.NewEncoder(writer).Encode(usrs)

}

// displays all awailable books - books that have quantity bigger than 0
func displayAwaliableBooks(writer http.ResponseWriter, request *http.Request) {
	rows, err := db.Query(GetAllAvailableBooks)

	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}

	defer rows.Close()

	var books []Book
	for rows.Next() {
		var book Book
		if err := rows.Scan(&book.ID, &book.BookName, &book.BookQuantity); err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}
		if book.BookQuantity > 0 {
			books = append(books, book)
		}
	}
	if len(books) == 0 {
		http.Error(writer, "No books are available", http.StatusConflict)
		return
	}

	json.NewEncoder(writer).Encode(books)
}

// add functions
// adds user with user_name and last_name
func addUser(writer http.ResponseWriter, request *http.Request) {
	var usr User

	if err := json.NewDecoder(request.Body).Decode(&usr); err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	if err := db.QueryRow(AddUser, usr.FirstName, usr.LastName).Scan(&usr.ID); err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusCreated)
	json.NewEncoder(writer).Encode(usr)
}

// borrows book based on book id if book quantitty is more than 0 and the book is in the db
func borrowBook(writer http.ResponseWriter, request *http.Request) {
	if err := json.NewDecoder(request.Body).Decode(&BorrowData); err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	var quantity int
	if err := db.QueryRow(GetBookQuantity, BorrowData.BookID).Scan(&quantity); err != nil {
		http.Error(writer, "Book not found", http.StatusNotFound)
		return
	}

	if quantity <= 0 {
		http.Error(writer, "Book not available", http.StatusConflict)
		return
	}

	_, err := db.Exec(BorrowBook, BorrowData.UserID, BorrowData.BookID)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = db.Exec(DecreaseBookQty, BorrowData.BookID)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusOK)
	json.NewEncoder(writer).Encode(map[string]string{"status": "book borrowed"})
}

// returns book based on book id if book entry exists
func returnBook(writer http.ResponseWriter, request *http.Request) {
	if err := json.NewDecoder(request.Body).Decode(&BorrowData); err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	res, err := db.Exec(ReturnBook, BorrowData.UserID, BorrowData.BookID)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	rows, err := res.RowsAffected()
	if err != nil || rows == 0 {
		http.Error(writer, "Borrowed book not in system", http.StatusNotFound)
		return
	}

	_, err = db.Exec(IncreaseBookQty, BorrowData.BookID)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusOK)
	json.NewEncoder(writer).Encode(map[string]string{"status": "book returned"})

}
