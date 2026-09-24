package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	_ "modernc.org/sqlite"
)

var db *sql.DB

func initDB() {
	var err error
	db, err = sql.Open("sqlite", ":memory:")
	if err != nil {
		log.Fatal(err)
	}
	db.Exec(`CREATE TABLE users (
		id   INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		email TEXT NOT NULL UNIQUE
	)`)
	db.Exec(`CREATE TABLE orders (
		id      INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL REFERENCES users(id),
		item    TEXT NOT NULL,
		amount  REAL NOT NULL
	)`)
	// Seed
	db.Exec(`INSERT INTO users(name,email) VALUES ('Alice','alice@example.com'),('Bob','bob@example.com')`)
	db.Exec(`INSERT INTO orders(user_id,item,amount) VALUES (1,'Widget',9.99),(1,'Gadget',24.99),(2,'Gizmo',14.99)`)
}

func main() {
	initDB()

	// INSERT user
	http.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			name := r.URL.Query().Get("name")
			email := r.URL.Query().Get("email")
			res, err := db.Exec("INSERT INTO users(name,email) VALUES (?,?)", name, email)
			if err != nil {
				http.Error(w, err.Error(), 400)
				return
			}
			id, _ := res.LastInsertId()
			fmt.Fprintf(w, `{"id":%d,"name":%q,"email":%q}`+"\n", id, name, email)
			return
		}
		// SELECT all users
		rows, err := db.Query("SELECT id,name,email FROM users")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer rows.Close()
		var users []map[string]any
		for rows.Next() {
			var id int64
			var name, email string
			rows.Scan(&id, &name, &email)
			users = append(users, map[string]any{"id": id, "name": name, "email": email})
		}
		json.NewEncoder(w).Encode(users)
	})

	// JOIN: user + their orders
	http.HandleFunc("/users/orders", func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query(`
			SELECT u.id, u.name, o.item, o.amount
			FROM users u
			JOIN orders o ON o.user_id = u.id
			ORDER BY u.id`)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer rows.Close()
		var results []map[string]any
		for rows.Next() {
			var uid int64
			var name, item string
			var amount float64
			rows.Scan(&uid, &name, &item, &amount)
			results = append(results, map[string]any{
				"user_id": uid, "name": name, "item": item, "amount": amount,
			})
		}
		w.Header().Set("X-Query-Type", "JOIN")
		json.NewEncoder(w).Encode(results)
	})

	log.Println("SQL service on :8080")
	http.ListenAndServe(":8080", nil)
}
