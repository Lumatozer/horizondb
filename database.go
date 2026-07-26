package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

type Response struct {
	Value string
	Ok    bool
}

func getConn(name string) *sql.DB {

    os.MkdirAll("databases", 0755)

    db, err := sql.Open("sqlite3", "databases/"+name)

    if err != nil {

        fmt.Println(err)

		return nil

    }



    query1 := "CREATE TABLE IF NOT EXISTS main (x TEXT PRIMARY KEY, y TEXT)"

    _, err = db.Exec(query1)

    if err != nil {

        fmt.Println(err)

		return nil

    }

    return db

}



func Set(db *sql.DB, key string, val string) {

    queryCheck := "SELECT x FROM main WHERE x = ?"

    var existingKey string

    err := db.QueryRow(queryCheck, key).Scan(&existingKey)



    if err == sql.ErrNoRows {

        queryInsert := "INSERT INTO main (x, y) VALUES (?, ?)"

        _, err = db.Exec(queryInsert, key, val)

        if err != nil {

            fmt.Println(err)

			return

        }

    } else if err != nil {

		fmt.Println(err)

		return

    } else {

        queryUpdate := "UPDATE main SET y = ? WHERE x = ?"

        _, err = db.Exec(queryUpdate, val, key)

        if err != nil {

            fmt.Println(err)

			return

        }

    }

}



func Get(db *sql.DB, key string) (string, bool) {

    query := "SELECT y FROM main WHERE x = ?"

    var value string

    err := db.QueryRow(query, key).Scan(&value)

    if err == sql.ErrNoRows {

        return "", false

    } else if err != nil {

		fmt.Println(err)

		return "", false

    }

    return value, true

}



func Get_All(db *sql.DB, key string) (map[string]string, bool) {

    query := "SELECT x, y FROM main WHERE x LIKE ?"

    rows, err := db.Query(query, key+"%")

    if err != nil {

        fmt.Println(err)

        return nil, false

    }

    defer rows.Close()



    values := make(map[string]string)

    for rows.Next() {

        var k, v string

        if err := rows.Scan(&k, &v); err != nil {

            fmt.Println(err)

            return nil, false

        }

        values[k] = v

    }

    if err := rows.Err(); err != nil {

        fmt.Println(err)

        return nil, false

    }



    if len(values) == 0 {

        return nil, false

    }

    return values, true

}



func Delete(db *sql.DB, key string) bool {

    query := "DELETE FROM main WHERE x = ?"

    result, err := db.Exec(query, key)

    if err != nil {

        fmt.Println(err)

        return false

    }

    rowsAffected, err := result.RowsAffected()

    if err != nil {

        fmt.Println(err)

        return false

    }

    return rowsAffected > 0

}

func main() {
	db := getConn("main.db")
        if db == nil {return}
	defer db.Close()

	var dbMutex sync.Mutex

	http.HandleFunc("/set", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		val := r.URL.Query().Get("val")
		if key == "" || val == "" {
			w.Write([]byte("error"))
			return
		}

		dbMutex.Lock()
		Set(db, key, val)
		dbMutex.Unlock()

		w.Write([]byte("true"))
	})

	http.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		if key == "" {
			return
		}

		dbMutex.Lock()
		value, ok := Get(db, key)
		dbMutex.Unlock()

		json.NewEncoder(w).Encode(Response{
			Value: value,
			Ok:    ok,
		})
	})

	http.HandleFunc("/get_all", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		if key == "" {
			return
		}

		dbMutex.Lock()
		value, ok := Get_All(db, key)
		dbMutex.Unlock()

		out, _ := json.Marshal(value)
		json.NewEncoder(w).Encode(Response{
			Value: string(out),
			Ok:    ok,
		})
	})

	http.HandleFunc("/delete", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		if key == "" {
			return
		}

		dbMutex.Lock()
		ok := Delete(db, key)
		dbMutex.Unlock()

		json.NewEncoder(w).Encode(Response{
			Ok: ok,
		})
	})

        fmt.Println("starting server on port 8249")
	fmt.Println(http.ListenAndServe(":8249", nil))
}
