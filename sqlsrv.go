package main
import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	_ "github.com/go-sql-driver/mysql" 
	"github.com/gorilla/mux"
)
const (
	dbDriver   = "mysql"
	dbUser     = "user"
	dbPassword = "password"
	dbHost     = "localhost"
	dbPort     = "3306"
	dbName     = "your_database"
)
const bearerToken = "your_bearer_token"
func main() {
	
	dbURI := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbUser, dbPassword, dbHost, dbPort, dbName)
	
	db, err := sql.Open(dbDriver, dbURI)
	if err != nil {
		log.Fatal("Error connecting to database:", err)
	}
	defer db.Close()
	
	router := mux.NewRouter()
	
	router.HandleFunc("/{table}/{id}", authMiddleware(handleDatabaseRequest(db))).Methods(http.MethodPut, http.MethodPatch, http.MethodDelete)
	router.HandleFunc("/{table}", authMiddleware(handleDatabaseRequest(db))).Methods(http.MethodPost)
	
	server := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	log.Println("Server listening on :8080")
	
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Error starting server:", err)
		}
	}()
	
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := server.Shutdown(ctx); err != nil {
		log.Println("Error shutting down server:", err)
	} else {
		log.Println("Server gracefully stopped")
	}
}
func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header missing", http.StatusUnauthorized)
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString != bearerToken {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}
func handleDatabaseRequest(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		table := vars["table"]
		id := vars["id"]
		switch r.Method {
		case http.MethodPost:
			go handleInsert(w, r, db, table) 
		case http.MethodPut:
			go handleUpdate(w, r, db, table, id) 
		case http.MethodPatch:
			go handlePatch(w, r, db, table, id) 
		case http.MethodDelete:
			go handleDelete(w, r, db, table, id) 
		default:
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		}
	}
}
func handleInsert(w http.ResponseWriter, r *http.Request, db *sql.DB, table string) {
	
	var data map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}
	
	columns := make([]string, 0, len(data))
	values := make([]interface{}, 0, len(data))
	for k, v := range data {
		columns = append(columns, k)
		values = append(values, v)
	}
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", table, strings.Join(columns, ", "), "?"+strings.Repeat(", ?", len(columns)-1))
	
	go func() {
		result, err := db.ExecContext(r.Context(), query, values...) 
		if err != nil {
			http.Error(w, fmt.Sprintf("Error inserting data: %v", err), http.StatusInternalServerError)
			return
		}
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			http.Error(w, fmt.Sprintf("Error getting affected rows: %v", err), http.StatusInternalServerError)
			return
		}
		w.Write([]byte(fmt.Sprintf("%d row(s) inserted", rowsAffected)))
	}()
}
func handleUpdate(w http.ResponseWriter, r *http.Request, db *sql.DB, table, id string) {
	
	var data map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}
	
	updates := make([]string, 0, len(data))
	values := make([]interface{}, 0, len(data))
	for k, v := range data {
		updates = append(updates, fmt.Sprintf("%s = ?", k))
		values = append(values, v)
	}
	query := fmt.Sprintf("UPDATE %s SET %s WHERE id = ?", table, strings.Join(updates, ", "))
	values = append(values, id) 
	
	go func() {
		result, err := db.ExecContext(r.Context(), query, values...)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error updating data: %v", err), http.StatusInternalServerError)
			return
		}
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			http.Error(w, fmt.Sprintf("Error getting affected rows: %v", err), http.StatusInternalServerError)
			return
		}
		w.Write([]byte(fmt.Sprintf("%d row(s) updated", rowsAffected)))
	}()
}
func handlePatch(w http.ResponseWriter, r *http.Request, db *sql.DB, table, id string) {
	
	var data map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}
	
	updates := make([]string, 0, len(data))
	values := make([]interface{}, 0, len(data))
	for k, v := range data {
		updates = append(updates, fmt.Sprintf("%s = ?", k))
		values = append(values, v)
	}
	query := fmt.Sprintf("UPDATE %s SET %s WHERE id = ?", table, strings.Join(updates, ", "))
	values = append(values, id) 
	
	go func() {
		result, err := db.ExecContext(r.Context(), query, values...)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error patching data: %v", err), http.StatusInternalServerError)
			return
		}
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			http.Error(w, fmt.Sprintf("Error getting affected rows: %v", err), http.StatusInternalServerError)
			return
		}
		w.Write([]byte(fmt.Sprintf("%d row(s) patched", rowsAffected)))
	}()
}
func handleDelete(w http.ResponseWriter, r *http.Request, db *sql.DB, table, id string) {
	
	query := fmt.Sprintf("DELETE FROM %s WHERE id = ?", table)
	
	go func() {
		result, err := db.ExecContext(r.Context(), query, id)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error deleting data: %v", err), http.StatusInternalServerError)
			return
		}
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			http.Error(w, fmt.Sprintf("Error getting affected rows: %v", err), http.StatusInternalServerError)
			return
		}
		w.Write([]byte(fmt.Sprintf("%d row(s) deleted", rowsAffected)))
	}()
}
