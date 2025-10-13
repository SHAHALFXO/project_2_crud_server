package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"unicode"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var db *sql.DB

type User struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Password string `json:"password"`
	Email    string `json:"email"`
}
type Loginrequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
	user := os.Getenv("user")
	password := os.Getenv("password")
	dbname := os.Getenv("dbname")

	connstring :=fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable",user,password,dbname)
	db, err = sql.Open("postgres", connstring)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println("connected....")
	}

}

func passvalidation(pass string) error {

	if len(pass) < 8 {
		return errors.New("password must contain 8 characters")
	}

	var hasupper, haslower, hasnumber bool

	for _, char := range pass {
		switch {
		case unicode.IsUpper(char):
			hasupper = true
		case unicode.IsLower(char):
			haslower = true
		case unicode.IsNumber(char):
			hasnumber = true

		}
	}
	if !haslower || !hasupper || !hasnumber {
		return errors.New("password must contain atleast one uppercase,lowercase and a number")

	}

	return nil
}

func createuser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "only post method is allowed!!!", http.StatusMethodNotAllowed)
		return
	}
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w,"decoding. failed",http.StatusBadRequest)
		return
	}

	if user.Name == "" {
		http.Error(w, "Name required", http.StatusBadRequest)
		return
	}
	if user.Password == "" {
		http.Error(w, "password required", http.StatusBadRequest)
		return
	} else {
		if err := passvalidation(user.Password); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	if user.Email == "" {
		http.Error(w, "Email required", http.StatusBadRequest)
		return
	}
	if !strings.Contains(user.Email, "@") || !strings.Contains(user.Email, ".") {
		http.Error(w, "invalid format", http.StatusBadRequest)
		return
	}

	err = db.QueryRow("insert into users (name,password,email)values($1,$2,$3) Returning id", user.Name, user.Password, user.Email).Scan(&user.ID)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			http.Error(w, "Email already Exists", http.StatusConflict)
			return
		}
		log.Printf("inserting failed: %v", err)
		http.Error(w, "something went wrong in creation", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)

	fmt.Fprintf(w, "user with id %d created", user.ID)
}

func readuserbyid(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "only get method will accept", http.StatusBadRequest)
		return
	}

	idstring := strings.TrimPrefix(r.URL.Path, "/users/")
	if idstring == "" {
		http.Error(w, "User id required", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idstring)
	if err != nil {
		http.Error(w, "string conversion failed", http.StatusBadRequest)
		return
	}
	var user User
	err = db.QueryRow("select id,name,password,email from users where id =$1", id).Scan(&user.ID, &user.Name, &user.Password, &user.Email)
	if err != nil {
		http.Error(w, "details collection failed from database", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)

}

func getusers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "only get method will accept", http.StatusBadRequest)
		return
	}

	row, err := db.Query("select id,name,password,email from users")
	if err != nil {
		http.Error(w, "Query incorrect or failed", http.StatusInternalServerError)
	}

	defer row.Close()
	var users []map[string]interface{}

	for row.Next() {
		var id int
		var name, password, email string
		row.Scan(&id, &name, &password, &email)

		users = append(users, map[string]interface{}{
			"id": id, "name": name, "password": password, "email": email,
		})
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)

}

func updateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "only put method allowed!!!", http.StatusBadRequest)
		return
	}

	var user User

	idstr := strings.TrimPrefix(r.URL.Path, "/users/")

	id, _ := strconv.Atoi(idstr)

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if user.Name == "" {
		http.Error(w, "Name required", http.StatusBadRequest)
		return
	}
	if user.Password == "" {
		http.Error(w, "password required", http.StatusBadRequest)
		return
	} else {
		if err := passvalidation(user.Password); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	if user.Email == "" {
		http.Error(w, "Email required", http.StatusBadRequest)
		return
	}
	if !strings.Contains(user.Email, "@") || !strings.Contains(user.Email, ".") {
		http.Error(w, "invalid format", http.StatusBadRequest)
		return
	}


	res, err := db.Exec("update users set name=$1,email=$2,password=$3 where id=$4", user.Name, user.Email, user.Password, id)
	if err != nil {
		http.Error(w, "Updation failed", http.StatusInternalServerError)
		return
	}
	
	rowsAffected,_:=res.RowsAffected()
     if rowsAffected==0{
		http.Error(w,"User not found",http.StatusNotFound)
		return
	 }


	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "updated succesfully")

}

func deleteUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "This method not allowed", http.StatusBadRequest)
		return
	}
	idstr := strings.TrimPrefix(r.URL.Path, "/users/")
	id, _ := strconv.Atoi(idstr)

	res, err := db.Exec("delete from users where id=$1", id)
	if err != nil {
		http.Error(w, "Deletion failed", http.StatusInternalServerError)
		return
	}

	rowsAffected,_:=res.RowsAffected()
     if rowsAffected==0{
		http.Error(w,"User not found",http.StatusNotFound)
		return
	 }

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "deleted succesfully")
}

func login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "This method is not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req Loginrequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "decoding failed", http.StatusInternalServerError)
		return
	}

	if req.Email == "shahal@gmail.com" && req.Password == "0987" {
		token := "12345"
		resp := map[string]string{"Message": "Login succesful", "Token": token}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	} else {
		http.Error(w, "Login failed", http.StatusBadRequest)
		return
	}

}

func main() {
	http.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			getusers(w, r)
		case http.MethodPost:
			createuser(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte("this method not allowed"))
		}

	})
	http.HandleFunc("/users/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			readuserbyid(w, r)
		case http.MethodPut:
			updateUser(w, r)
		case http.MethodDelete:
			deleteUser(w, r)
		default:
			http.Error(w, "this method not allowed", http.StatusBadRequest)
			return
		}
	})

	http.HandleFunc("/login", login)

	fmt.Println("server running")
	http.ListenAndServe(":8080", nil)

}
