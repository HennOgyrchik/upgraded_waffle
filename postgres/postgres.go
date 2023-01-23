package postgres

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)

func dbConnection() (*sql.DB, error) {
	connStr := "user=root password=waffel dbname=postgres sslmode=disable host=postgres port=5432"
	db, err := sql.Open("postgres", connStr)

	if err != nil {
		fmt.Printf("Ошибка подключения к БД:%s\n", err)
		db.Close()
	}

	return db, err
}

func Registration(login string, pass string) error {
	db, err := dbConnection()
	if err != nil {
		return err
	}
	stmt, err := db.Prepare("insert into users values ($1,$2)")
	if err != nil {
		return err
	}
	_ = stmt.QueryRow(login, pass)
	defer db.Close()
	return err
}

func Authorization(login string, pass string) (bool, error) {
	db, err := dbConnection()
	if err != nil {
		return false, err
	}
	defer db.Close()

	ok, err := CheckLogin(login)
	if err != nil {
		return false, err
	}
	if ok {
		return false, err
	}

	stmt, err := db.Prepare("select pass from users where login=$1")
	if err != nil {
		return false, err
	}
	var temp string
	err = stmt.QueryRow(login).Scan(&temp)
	if err != nil {
		return false, err
	}
	if pass == temp {
		return true, err
	} else {
		return false, err
	}
}

func CheckLogin(login string) (bool, error) { //возвращает true, если логин НЕ найден в бд
	db, err := dbConnection()
	if err != nil {
		return false, err
	}
	defer db.Close()

	stmt, err := db.Prepare("select count(login) from users where login=$1")
	if err != nil {
		fmt.Println("!!!!!!!!!!!!1", err)
		return false, err
	}
	var temp int
	err = stmt.QueryRow(login).Scan(&temp)
	if err != nil {
		return false, err
	}
	if temp > 0 {
		return false, err
	} else {
		return true, err
	}

}

func WriteMessage(login, mes string) error {
	db, err := dbConnection()
	if err != nil {
		return err
	}
	defer db.Close()
	stmt, err := db.Prepare("insert into messages values (default,$1,$2)")
	if err != nil {
		return err
	}
	_ = stmt.QueryRow(login, mes)
	defer db.Close()
	return err

}

func GetLastMessage() (string, error) {
	db, err := dbConnection()
	if err != nil {
		return "", err
	}
	defer db.Close()

	row := db.QueryRow(" select login, message from messages order by id desc limit 1")
	var login, message string
	err = row.Scan(&login, &message)
	if err != nil {
		return "", err
	}
	return fmt.Sprint(login, ": ", message, "\n"), err
}
