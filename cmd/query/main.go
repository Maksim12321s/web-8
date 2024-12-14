package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	_ "github.com/lib/pq"
)

type DatabaseProvider struct {
	db *sql.DB
}
type Handlers struct {
	dp DatabaseProvider
}

func (dp *DatabaseProvider) Insert_hello(msg string) (err error) {
	_, err = dp.db.Exec(`insert into hello (message) values ($1)`, msg)
	return err

}
func (dp *DatabaseProvider) Get_from_hello(id int) (msg string, err error) {
	row := dp.db.QueryRow(`select message from hello where id = $1;`, id)
	if err != nil {
		return "", err
	}
	if err = row.Scan(&msg); err != nil {
		return "", err
	}
	return
}
func (dp *DatabaseProvider) Get_all_from_hello() (arr []string, err error) {
	row, err := dp.db.Query("select message from hello;")
	if err != nil {
		return nil, err
	}
	var val string
	for row.Next() {
		if err = row.Scan(&val); err != nil {
			return nil, err
		}
		arr = append(arr, val)
	}

	return arr, nil

}

type Post struct {
	Id  *int    `id:"id"`
	Msg *string `json:"msg"`
}

func (h *Handlers) handler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":

		arr, err := h.dp.Get_all_from_hello()
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
		}
		us := struct {
			Arr []string `json:"hello"`
		}{}
		us.Arr = arr
		u, err := json.Marshal(&us)
		if err != nil {
			log.Fatal(err)
		}
		w.Write(u)
	case "POST":
		post := Post{}
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&post); err != nil {
			log.Fatal(err.Error())
		}
		if post.Id == nil && post.Msg == nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("No data"))
		} else if post.Id == nil {
			if err := h.dp.Insert_hello(*post.Msg); err != nil {
				w.Write([]byte(err.Error()))
				return
			}
			w.Write([]byte("Written"))
		} else {
			r := struct {
				Msg string `json:"msg"`
			}{}
			r.Msg, _ = h.dp.Get_from_hello(*post.Id)
			rez, _ := json.Marshal(&r)
			w.Write(rez)

		}

	}
}

const (
	dbname   = "sandbox"
	port     = 5432
	host     = "localhost"
	user     = "maksim"
	password = "123"
)

func main() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer db.Close()
	dp := DatabaseProvider{db: db}
	h := Handlers{dp: dp}
	http.HandleFunc("/", h.handler)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err.Error())
	}
}
