package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	_ "github.com/redis/go-redis/v9"
	"gitlab.com/idoko/rediboard/db"

	"io"
	"net/http"
	"strconv"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "123123"
	dbname   = "postgres"
)

const (
	hostR = "127.0.0.1"
	portR = 6379
)

type Database struct {
	Client *redis.Client
}

var ab *sql.DB
var redisD *db.Database
var rdc *redis.Client
var ctx = context.TODO()

func main() {
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	var err error
	ab, err = sql.Open("postgres", psqlconn)
	if err != nil {
		fmt.Printf("ERROR : %s ", err)
		return
	}
	defer ab.Close()

	ec := echo.New()
	ec.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			fmt.Printf("%v", c.Request())
			if err = next(c); err != nil {
				c.Error(err)
			}
			return nil
		}
	})
	//sharding :
	//redis :
	redisD, err = db.NewDatabase("localhost:6379")
	if err != nil {
		fmt.Printf("ERROR : %s ", err)
		return
	}

	ec.GET("/GET", GET)
	ec.PUT("/UPDATE", UPDATE)
	ec.PUT("/INSERT", POST)
	ec.DELETE("/DELETE", DELETE)
	ec.Logger.Fatal(ec.Start("localhost:8081"))

}

func DELETE(c echo.Context) error {
	id := c.QueryParams().Get("id")
	code, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return err
	}
	_, err = ab.Exec("DELETE from httpserver where code = $1", code)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return echo.NewHTTPError(http.StatusNotFound, "Invalid id")
		}
		fmt.Printf("ERROR: %v", err)
		return err
	}
	result := fmt.Sprintf("your note successfuly deleted ")
	return c.String(http.StatusOK, result)
}
func UPDATE(c echo.Context) error {
	var context []byte
	id := c.QueryParams().Get("id")
	code, err := strconv.ParseInt(id, 10, 64)
	context, err = io.ReadAll(c.Request().Body)
	if err != nil {
		return err
	}
	newNote := string(context)
	_, err = ab.Exec("UPDATE httpserver set note = $1 where code = $2", &newNote, &code)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return echo.NewHTTPError(http.StatusNotFound, "Invalid id")
		}
		fmt.Printf("ERROR: %v", err)
		return err
	}
	result := fmt.Sprintf("your newNote successfuly saved")
	return c.String(http.StatusOK, result)
}

func GET(c echo.Context) error {
	var note string
	id := c.QueryParams().Get("id")
	code, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		fmt.Printf("ERROR: %v", err)
		return err
	}
	// if we want to save notes in redis  that repeated most of 2 time :
	KeyHLL := "key" + id
	result, err := rdc.PFCount(ctx, KeyHLL).Result()
	if err != nil {
		return err
	}
	if result > 2 {
		resultNote := rdc.Get(ctx, id)
		fmt.Printf("your codes note equal to : %s", resultNote)
		return c.String(http.StatusOK, resultNote.Val())
	} else if result < 2 {
		result++
		err = rdc.PFAdd(ctx, KeyHLL, result).Err()
		if err != nil {
			return err
		}
	}
	err = ab.QueryRow("SELECT note FROM httpserver WHERE code = $1", int(code)).Scan(&note)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return echo.NewHTTPError(http.StatusNotFound, "Invalid id")
		}
		fmt.Printf("ERROR: %v", err)
		return err
	}
	if result == 2 {
		err = rdc.Set(ctx, id, note, 0).Err()
		if err != nil {
			return err
		}
	}
	fmt.Printf("your codes note equal to : %s", note)
	return c.String(http.StatusOK, note)
}

func POST(c echo.Context) error {
	var codeNote int
	content, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return err
	}
	note := string(content)
	err = ab.QueryRow("INSERT INTO httpserver (note)VALUES ($1) RETURNING code", note).
		Scan(&codeNote)
	if err != nil {
		return err
	}
	response := fmt.Sprintf("your codes note equal to : %d", codeNote)
	return c.String(http.StatusOK, response)
}

//func UPDATE(c echo.Context)error  {
//e.PUT("/employee", func(c echo.Context) error {
//	if err := c.Bind(); err != nil {
//		return err
//	}
//	sqlStatement := "UPDATE employees SET name=$1,salary=$2,age=$3 WHERE id=$5"
//	res, err := ab.Query(sqlStatement, u.Name, u.Salary, u.Age, u.Id)
//	if err != nil {
//		fmt.Println(err)
//		//return c.JSON(http.StatusCreated, u);
//	} else {
//		fmt.Println(res)
//		return c.JSON(http.StatusCreated, u)
//	}
//	return c.String(http.StatusOK, u.Id)
//})
//}
