package database

import (
	"database/sql"
	"fmt"
	"log"
	"project3/config"
	"project3/models"

	"github.com/sirupsen/logrus"

	_ "github.com/go-sql-driver/mysql"
)

func ConnecToMysql() (*sql.DB, error) {
	dbSql, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", config.ViperConfig.MysqlUser, config.ViperConfig.MysqlPassword, config.ViperConfig.MysqlHost, config.ViperConfig.MysqlPort, config.ViperConfig.MysqlDBName))
	if err != nil {
		logrus.Errorln("Error connecting to mysql ", err)
		return nil, err
	}

	if err = dbSql.Ping(); err != nil {
		logrus.Errorln("Error while pinging to mysql ", err)
		return nil, err
	}
	return dbSql, nil
}

func CheckUserIfExist(dbSql *sql.DB, email string) (bool, error) {
	var userid string
	if err := dbSql.QueryRow(fmt.Sprintf(`SELECT userid from users where emailid='%s'`, email)).Scan(&userid); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	if userid == "" {
		return false, nil
	}

	return true, nil
}

func AddUser(dbSql *sql.DB, users models.User) error {
	st := fmt.Sprintf(`INSERT INTO users (userid, emailid, firstname, lastname, password) values('%s','%s','%s','%s','%s')`, users.UserId, users.Email, users.FirstName, users.LastName, users.Password)
	log.Println(st)
	_, err := dbSql.Query(st)

	if err != nil {
		return err
	}
	return nil
}

func GetUserByEmail(dbSql *sql.DB, emailId string) (models.User, error) {
	var user models.User
	st := fmt.Sprintf(`SELECT userid, emailid, firstname, lastname, password from users where emailid='%s'`, emailId)
	if err := dbSql.QueryRow(st).Scan(&user.UserId, &user.Email, &user.FirstName, &user.LastName, &user.Password); err != nil {
		return user, err
	}
	return user, nil
}

func InsertBook(dbSql *sql.DB, book models.Book) error {
	st := fmt.Sprintf(`INSERT INTO books (bookid, isbn, title, genre, authorname, authorcountry) values('%s','%s','%s','%s','%s', '%s')`, book.ID, book.Isbn, book.Title, book.Genre, book.Author.Name, book.Author.Country)
	_, err := dbSql.Query(st)
	if err != nil {
		return err
	}
	return nil
}

func GetBook(dbSql *sql.DB, bookid string) (models.Book, error) {
	var book models.Book
	st := fmt.Sprintf("SELECT bookid, isbn, title, genre, authorname, authorcountry from books where bookid='%s'", bookid)
	if err := dbSql.QueryRow(st).Scan(&book.ID, &book.Isbn, &book.Title, &book.Genre, &book.Author.Name, &book.Author.Country); err != nil {
		return book, err
	}
	return book, nil
}

func GetAllBooksByKey(dbSql *sql.DB) ([]models.Book, error) {
	var books []models.Book
	fmt.Printf("dbSql.Ping(): %v\n", dbSql.Ping())
	st := fmt.Sprintf("SELECT bookid, isbn, title, genre, authorname, authorcountry from books")
	result, err := dbSql.Query(st)
	if err != nil {
		return books, err
	}
	for result.Next() {
		var book models.Book
		if result.Scan(&book.ID, &book.Isbn, &book.Title, &book.Genre, &book.Author.Name, &book.Author.Country); err != nil {
			return books, err
		}
		books = append(books, book)
	}
	return books, nil
}

func UpdateBook(dbSql *sql.DB, id string, book models.Book) error {
	st := fmt.Sprintf(`UPDATE books set isbn = '%s', title = '%s', genre = '%s', authorname = '%s', authorcountry = '%s' where bookid = '%s'`, book.Isbn, book.Title, book.Genre, book.Author.Name, book.Author.Country, id)
	_, err := dbSql.Query(st)
	if err != nil {
		return err
	}
	return nil
}

func DeleteBook(dbSql *sql.DB, id string) error {
	st := fmt.Sprintf(`DELETE FROM books where bookid = '%s'`, id)
	_, err := dbSql.Query(st)
	if err != nil {
		return err
	}
	return nil
}
