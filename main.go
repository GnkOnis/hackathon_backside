package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/oklog/ulid"
	"github.com/rs/cors"
	_ "github.com/rs/cors"
	"log"
	_ "math/rand"
	"net/http"
	"os"
	"time"
	_ "time"
)

type SendJson struct {
	Title      string `json:"title"`
	Category   int    `json:"category"`
	Curr       int    `json:"curr"`
	Link       string `json:"link"`
	CreateTime string `json:"createtime"`
	UpdateTime string `json:"updatetime"`
	Summary    string `json:"summary"`
	Name       string `json:"name"`
}

var db *sql.DB

func init() {
	//デプロイ用
	mysqlUser := os.Getenv("MYSQL_USER")
	mysqlPwd := os.Getenv("MYSQL_PWD")
	mysqlHost := os.Getenv("MYSQL_HOST")
	mysqlDatabase := os.Getenv("MYSQL_DATABASE")
	connStr := fmt.Sprintf("%s:%s@%s/%s", mysqlUser, mysqlPwd, mysqlHost, mysqlDatabase)
	_db, err := sql.Open("mysql", connStr)

	////mysqlのコンテナに接続
	//sqluser := "test_user"
	//sqlpwd := "password"
	//sqldatabase := "test_database"
	//_db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@(localhost:3306)/%s", sqluser, sqlpwd, sqldatabase))

	if err != nil {
		log.Fatalf("fail: sql.Open, %v\n", err)
	}
	if err := _db.Ping(); err != nil {
		log.Fatalf("fail: _db.Ping, %v\n", err)
	}
	db = _db
}

func handler_table(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		order := r.URL.Query().Get("order")
		category := r.URL.Query().Get("category")
		curr := r.URL.Query().Get("curr")

		var categoryStr string
		if category == "0" {
			categoryStr = "1,2,3,"
		} else {
			categoryStr = category
		}
		var currStr string
		if curr == "0" {
			currStr = "1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27"
		} else {
			currStr = curr
		}

		var orderStr string
		switch order {
		case "0":
			orderStr = " ORDER BY createtime DESC"
		case "1":
			orderStr = " ORDER BY createtime ASC"
		case "2":
			orderStr = " ORDER BY updatetime DESC"
		case "3":
			orderStr = " ORDER BY updatetime ASC"
		default:
			http.Error(w, "Invalid sort", http.StatusBadRequest)
			return
		}
		query := "SELECT * FROM maintable WHERE category IN (" + categoryStr + ") AND curr IN (" + currStr + ") " + orderStr

		rows, err := db.Query(query)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		items := []SendJson{}
		for rows.Next() {
			var item SendJson
			err := rows.Scan(&item.Title, &item.Category, &item.Curr, &item.Link, &item.CreateTime, &item.UpdateTime, &item.Summary, &item.Name)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			items = append(items, item)
		}
		response, err := json.Marshal(items)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(response)
	case http.MethodPost:
		var Recievejson struct {
			Title    string `json:"title"`
			Category int    `json:"category"`
			Curr     int    `json:"curr"`
			Link     string `json:"link"`
			Comment  string `json:"comment"`
			Name     string `json:"name"`
		}
		err := json.NewDecoder(r.Body).Decode(&Recievejson)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			fmt.Println("err := json.NewDecoder(r.Body).Decode(&inputData)でエラーが発生")
			return
		}
		nowTime := time.Now()
		query := "INSERT INTO maintable (title,category,curr,link,createtime,updatetime,comment,name) VALUES(?,?,?,?,?,?,?,?)"
		_, err = db.Exec(query, Recievejson.Title, Recievejson.Category, Recievejson.Curr, Recievejson.Link, nowTime, nowTime, Recievejson.Comment, Recievejson.Name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
	case http.MethodPut:
	case http.MethodDelete:
	}
}
func handler_element(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
	case http.MethodPost:
	case http.MethodDelete:
	}
}

func main() {
	http.HandleFunc("/table", handler_table)
	http.HandleFunc("/element", handler_element)

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
	})
	handlerWithCors := c.Handler(http.DefaultServeMux)
	log.Println("Listening...")
	if err := http.ListenAndServe(":8080", handlerWithCors); err != nil {
		log.Fatal(err)
	}
}
