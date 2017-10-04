package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"strings"
	"sync"

	_ "github.com/go-sql-driver/mysql"
)

var wg sync.WaitGroup

type Todo struct {
	id         int
	todo       string
	created_at string
	completed  int
}

func NewTodo(todo string, db *sql.DB) {
	fmt.Println(todo)
	add, err := db.Prepare("INSERT INTO todos(todo, created_at, completed) VALUES (? , NOW(), 0)")

	if err != nil {
		panic(err.Error())
	}
	defer add.Close()

	_, err = add.Exec(todo)

	if err != nil {
		panic(err.Error())
	}

	log.Printf("Todo '%s' added to your List.\n", todo)

	wg.Done()
}

func CompleteTodo(db *sql.DB, id int) {

	stmnt, err := db.Prepare("UPDATE todos SET completed = 1 WHERE id = ?")

	if err != nil {
		panic(err.Error())
	}

	_, err = stmnt.Exec(id)

	if err != nil {
		panic(err.Error())
	}

	log.Println("Todo Completed :)")

	wg.Done()
}

func ShowTodos(db *sql.DB, option string) {

	var todos *sql.Rows
	var err error

	switch option {
	case "sc":
		todos, err = db.Query("SELECT * FROM todos WHERE completed = 1")
		break
	case "st":
		todos, err = db.Query("SELECT * FROM todos WHERE completed = 0")
		break
	default:
		todos, err = db.Query("SELECT * FROM todos")
		break
	}

	if err != nil {
		panic(err.Error())
	}

	defer todos.Close()

	todoArr := make([]*Todo, 0)
	for todos.Next() {
		td := new(Todo)

		err := todos.Scan(&td.id, &td.todo, &td.created_at, &td.completed)

		if err != nil {
			log.Fatalln(err)
		}

		todoArr = append(todoArr, td)
	}

	if err = todos.Err(); err != nil {
		log.Fatalln(err)
	}

	for _, todo := range todoArr {
		if todo.completed == 0 {
			fmt.Println(todo.id, todo.todo, todo.created_at)
		} else {
			fmt.Println(todo.id, todo.todo, todo.created_at, "\u2713")
		}

	}

	wg.Done()
}

func main() {
	create := flag.Bool("n", false, "create new todo")
	complete := flag.Int("c", -1, "complete todo")
	show := flag.Bool("s", false, "show todo list")
	showcomplete := flag.Bool("sc", false, "show only completed")
	showtodo := flag.Bool("st", false, "show uncompleted")

	flag.Parse()
	db, err := sql.Open("mysql", "root:12345678@/godo")

	if err != nil {
		panic(err)
	}

	err = db.Ping()
	defer db.Close()
	if err != nil {
		panic("Could not reach Database")
	}

	fmt.Println("Connection to Database Established")

	fmt.Println()
	fmt.Println()

	if *create {
		wg.Add(1)
		go NewTodo(strings.Join(flag.Args(), " "), db)
	}

	if *complete != -1 {
		wg.Wait()
		wg.Add(1)
		go CompleteTodo(db, *complete)
	}
	if *show {
		wg.Wait()
		wg.Add(1)
		go ShowTodos(db, "")
	}
	if *showcomplete {
		wg.Wait()
		wg.Add(1)
		go ShowTodos(db, "sc")
	}
	if *showtodo {
		wg.Wait()
		wg.Add(1)
		go ShowTodos(db, "st")
	}
	wg.Wait()
}
