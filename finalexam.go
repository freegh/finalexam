package main

import (
	"fmt"
	"database/sql"
	"log"
	_"github.com/lib/pq"
	"os"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Customer struct{
	ID int `json:"id"`
	Name string `json:"name"`
	Email string `json:"email"`
	Status string `json:"status"`
}

var db *sql.DB
var err error
var customers []Customer
var t Customer

func authMidleware(c *gin.Context)  {
	fmt.Println("GoodMorning middleware")
	token := c.GetHeader("Authorization")
	if token != "token2019" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized."})
		c.Abort()
		return
	}
	c.Next()
	fmt.Println("after in middleware")
}

func connectDb() {
	url := os.Getenv("DATABASE_URL")
	fmt.Println("url: ", url)

	db, err = sql.Open("postgres", url)
	if err != nil {
		fmt.Println("Connect to database error", err)
	}
	fmt.Println("Already connect database")
}

func createDb()  {
	_, checkTb := db.Query("select * from customer;")
	if checkTb == nil {
		fmt.Println("table exist")
	}else{
		createDb := `
		CREATE TABLE IF NOT EXISTS customer (
			id SERIAL PRIMARY KEY,
			name TEXT,
			email TEXT,
			status TEXT
		);
		`
		if _, err = db.Exec(createDb); err != nil {
			log.Fatal("can't create table", err)
		}

		fmt.Println("Create table success")
	}
}

func addData(c *gin.Context)  {
	if err := c.ShouldBindJSON(&t); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	row := db.QueryRow("INSERT INTO customer (name, email, status) values ($1, $2, $3) RETURNING id", t.Name, t.Email, t.Status)
	if err = row.Scan(&t.ID); err != nil {
		fmt.Println("can't scan id", err)
		return
	}
	c.JSON(http.StatusCreated, t)
	fmt.Println("add data success", t)
}

func queryData(c *gin.Context)  {
	id := c.Param("id")
	
	stmt, err := db.Prepare("SELECT id, name, email, status FROM customer WHERE id=$1;")
	if err != nil {
		fmt.Println("can't prepare query")
	}
	row := stmt.QueryRow(id)
	err = row.Scan(&t.ID, &t.Name, &t.Email, &t.Status)
	if err != nil {
		fmt.Println("can't scan query ", err)
	}
	c.JSON(http.StatusOK,t)
	fmt.Println("query success", t)
}

func queryAllData(c *gin.Context)  {
	stmt ,err := db.Prepare("SELECT id, name, email, status FROM customer")
	if err != nil {
		fmt.Println("can't prepare query ", err)
	}
	rows, err := stmt.Query()
	if err != nil {
		fmt.Println("can't query ", err)
	}
	for rows.Next(){
		err := rows.Scan(&t.ID, &t.Name, &t.Email, &t.Status)
		if err != nil {
			fmt.Println("can't scan")
		}
		fmt.Println(t)
		customers = append(customers, t)
	}
	fmt.Println("query all success :", customers)
	c.JSON(http.StatusOK, customers)
}

func updateData(c *gin.Context)  {
	id := c.Param("id")

	stmt, err := db.Prepare("UPDATE customer SET name=$1, email=$2, status=$3 WHERE id=$4;")
	if err != nil {
		fmt.Println("can't prepare")
	}
	if err := c.ShouldBindJSON(&t); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if _, err := stmt.Exec(t.ID,t.Name, t.Email, t.Status, id); err != nil{
		fmt.Println("error execute")
	}
	fmt.Println("update success")
	c.JSON(http.StatusOK, t)
}

func deleteData(c *gin.Context)  {
	id := c.Param("id")
	stmt, err := db.Prepare("DELETE FROM customer WHERE id=$1")
	if err != nil {
		fmt.Println("can't prepare")
	}
	if _, err := stmt.Exec(id); err != nil {
		fmt.Println("error execute")
	}
	fmt.Println("Delete success")
	c.JSON(http.StatusOK, gin.H{"message":"customer deleted",})
}

func main()  {
	r := gin.Default()

	connectDb()
	createDb()
	r.Use(authMidleware)
	r.POST("/customers", addData)
	r.GET("/customers/:id", queryData)
	r.GET("/customers", queryAllData)
	r.PUT("customers/:id", updateData)
	r.DELETE("customers/:id", deleteData)
	r.Run(":2019")

	defer db.Close()
}