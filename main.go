package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/oklog/ulid"
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
	Id         string `json:"id"`
	Title      string `json:"title"`
	Category   int    `json:"category"`
	Curr       int    `json:"curr"`
	Link       string `json:"link"`
	CreateTime string `json:"createtime"`
	UpdateTime string `json:"updatetime"`
	Numcomment int    `json:"numcomment"`
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

	//mysqlのコンテナに接続
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
			categoryStr = "1,2,3"
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
		////////////////////////////////
		fmt.Println(query)
		//////////////////////////////
		rows, err := db.Query(query)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		items := []SendJson{}
		for rows.Next() {
			var item SendJson
			err := rows.Scan(&item.Id, &item.Title, &item.Category, &item.Curr, &item.Link, &item.CreateTime, &item.UpdateTime, &item.Numcomment, &item.Summary, &item.Name)
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

		//ULIDを生成
		ulid, err := ulid.New(ulid.Now(), nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		ulidStr := ulid.String()

		query := "INSERT INTO maintable (id,title,category,curr,link,createtime,updatetime,numcomment,summary,name) VALUES(?,?,?,?,?,?,?,?,?,?)"
		_, err = db.Exec(query, ulidStr, Recievejson.Title, Recievejson.Category, Recievejson.Curr, Recievejson.Link, nowTime, nowTime, 1, "まだsummaryは実装されていません", Recievejson.Name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		//ここでコメントとコメント主を保存するtableを新しくつくる
		err = createCommentTableAndInsertData(db, ulidStr, Recievejson.Comment, Recievejson.Name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
	case http.MethodPut:
		var Recievejson struct {
			Id       string `json:"id"`
			Title    string `json:"title"`
			Category int    `json:"category"`
			Curr     int    `json:"curr"`
			Link     string `json:"link"`
			Name     string `json:"name"`
		}
		err := json.NewDecoder(r.Body).Decode(&Recievejson)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if Recievejson.Title != "nochange" {
			_, err = db.Exec("UPDATE maintable SET title = ? WHERE id = ?", Recievejson.Title, Recievejson.Id)
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if Recievejson.Category != 100 {
			_, err = db.Exec("UPDATE maintable SET category = ? WHERE id = ?", Recievejson.Category, Recievejson.Id)
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if Recievejson.Curr != 100 {
			_, err = db.Exec("UPDATE maintable SET curr = ? WHERE id = ?", Recievejson.Curr, Recievejson.Id)
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if Recievejson.Link != "nochange" {
			_, err = db.Exec("UPDATE maintable SET link = ? WHERE id = ?", Recievejson.Link, Recievejson.Id)
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if Recievejson.Name != "nochange" {
			_, err = db.Exec("UPDATE maintable SET name = ? WHERE id = ?", Recievejson.Name, Recievejson.Id)
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		currentTime := time.Now()
		_, err = db.Exec("UPDATE maintable SET updatetime = ? WHERE id = ?", currentTime, Recievejson.Id)

		w.WriteHeader(http.StatusNoContent)
	case http.MethodDelete:
		id := r.URL.Query().Get("id")
		if id == "" {
			http.Error(w, "Title is required", http.StatusBadRequest)
		}
		query := "DELETE FROM maintable WHERE id = ?"
		_, err := db.Exec(query, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func createCommentTableAndInsertData(db *sql.DB, parentId, comment, name string) error {
	// テーブル名を生成
	tableName := parentId

	// コメントテーブルを作成
	createTableSQL := fmt.Sprintf("CREATE TABLE %s (id VARCHAR(26), name VARCHAR(64), comment VARCHAR(1024))", tableName)
	_, err := db.Exec(createTableSQL)
	if err != nil {
		return err
	}

	// ulidを生成
	ulid, err := ulid.New(ulid.Now(), nil)
	if err != nil {
		return err
	}

	// データを挿入
	insertSQL := fmt.Sprintf("INSERT INTO %s (id, name, comment) VALUES (?, ?, ?)", tableName)
	_, err = db.Exec(insertSQL, ulid.String(), name, comment)
	if err != nil {
		return err
	}
	return nil
}

func handler_element(w http.ResponseWriter, r *http.Request) {
	parentId := r.URL.Query().Get("parent_id")
	switch r.Method {
	case http.MethodGet:
		fmt.Println("enter get of element")

		// テーブル名を生成
		tableName := parentId

		// データベースからデータを取得
		rows, err := db.Query("SELECT id, name, comment FROM " + tableName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		// 取得したデータを格納するスライス
		var items []map[string]interface{}

		for rows.Next() {
			var id, name, comment string
			if err := rows.Scan(&id, &name, &comment); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// データをマップに追加
			item := map[string]interface{}{
				"id":      id,
				"name":    name,
				"comment": comment,
			}
			items = append(items, item)
		}

		// データをJSON形式で返す
		responseJSON, err := json.Marshal(items)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(responseJSON)
	case http.MethodPost:
		// POSTリクエストの処理
		var Recievejson struct {
			Name    string `json:"name"`
			Comment string `json:"comment"`
		}
		err := json.NewDecoder(r.Body).Decode(&Recievejson)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// ULIDを生成
		ulid, err := ulid.New(ulid.Now(), rand.Reader)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// テーブル名を生成
		tableName := parentId

		// データベースにデータを挿入
		_, err = db.Exec("INSERT INTO "+tableName+" (id, name, comment) VALUES (?, ?, ?)", ulid.String(), Recievejson.Name, Recievejson.Comment)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
	case http.MethodDelete:
		id := r.URL.Query().Get("id")
		if id == "" {
			http.Error(w, "ID is required", http.StatusBadRequest)
			return
		}

		// テーブル名を生成
		tableName := parentId

		// データベースから指定されたIDのカラムを削除
		_, err := db.Exec("DELETE FROM "+tableName+" WHERE id = ?", id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
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
	if err := http.ListenAndServe(":8000", handlerWithCors); err != nil {
		log.Fatal(err)
	}
}
