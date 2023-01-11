package main

import (
	"database/sql"
	"fmt"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"io"
	"net/http"
	"strconv"
)

func server(w http.ResponseWriter, r *http.Request) {
	switch r.Method {

	case "DELETE":
		id := r.URL.Query().Get("id")
		code, _ := strconv.ParseInt(id, 0, 64)

		Delete := `delete from httpserver where code = $1`
		_, eer := ab.Exec(Delete, code)
		if eer != nil {
			fmt.Fprintf(w, "ERROR : %s ", eer)
			return
		} else {
			fmt.Fprintf(w, "your note successfully deleted .")
		}

	case "PUT":
		id := r.URL.Query().Get("id")
		code, _ := strconv.ParseInt(id, 0, 64)

		Newnote, err := io.ReadAll(r.Body)
		if err != nil {
			fmt.Fprintf(w, "%s : %s ", "could not read body", err)
			return
		}
		update := `Update httpserver set note = $1 where code = $2 `
		_, e := ab.Exec(update, Newnote, code)
		if e != nil {
			fmt.Fprintf(w, "ERROR : %s", e)
		} else {
			fmt.Fprintf(w, "your note updated ")
		}

	case "GET":
		id := r.URL.Query().Get("id")
		code, _ := strconv.ParseInt(id, 10, 0)
		var note string
		show := `SELECT note FROM httpserver WHERE code = $1 ;`
		er := ab.QueryRow(show, code).Scan(&note)
		if er != nil {
			if er == sql.ErrNoRows {
				fmt.Fprintf(w, "the note does not exist")
				return
			}
			fmt.Fprintf(w, "ERROR : %s", er)
			return
		} else {
			fmt.Fprintf(w, "your note in code %d is : %s ", code, note)
		}

	case "POST":
		body, err := io.ReadAll(r.Body)
		if err != nil {
			fmt.Fprintf(w, "%s : %s ", "could not read body", err)
			return
		}
		//insert note :
		insertNote := `insert into httpserver (note) values($1);`
		_, e := ab.Exec(insertNote, string(body))
		if e != nil {
			fmt.Fprintf(w, "ERROR :  '%s' ", e)
			return
		} else {
			var code string
			ee := ab.QueryRow(`SELECT code FROM httpserver WHERE note = $1 ;`, body).Scan(&code)
			if ee != nil {
				fmt.Fprintf(w, "ERROR : %s", ee)
				return
			}
			fmt.Fprintf(w, "your note successfully saved ")
			fmt.Fprintf(w, "your codes note equal to : %s", code)
		}
	}
}

var ab *sql.DB

const (
	host     = "localhost"
	port     = 5432
	user     = "notes_user"
	password = "123123"
	dbname   = "httpserver"
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", server)
	fmt.Println("server at port 8081 : ")

	// postgresql  :
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	var errP error
	ab, errP = sql.Open("postgres", psqlconn)
	if errP != nil {
		fmt.Printf("ERROR : %s ", errP)
	}
	defer ab.Close()
	errP = ab.Ping()
	if errP != nil {
		fmt.Printf("ERROR : %s ", errP)
	}

	if err := http.ListenAndServe("localhost:8081", r); err != nil {
		fmt.Println(err)
	}
}
# test
