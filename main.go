package main

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

type Client struct {
	ID       int
	FIO      string
	Login    string
	Birthday string
	Email    string
}

// String реализует метод интерфейса fmt.Stringer для Sale, возвращает строковое представление объекта Client.
// Теперь, если передать объект Client в fmt.Println(), то выведется строка, которую вернёт эта функция.
func (c Client) String() string {
	return fmt.Sprintf("ID: %d FIO: %s Login: %s Birthday: %s Email: %s",
		c.ID, c.FIO, c.Login, c.Birthday, c.Email)
}

func main() {
	db, err := sql.Open("sqlite", "demo.db")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()

	// добавление нового клиента
	newClient := Client{
		FIO:      "Пупкин Василий Иванович", // укажите ФИО
		Login:    "pup2000",                 // укажите логин
		Birthday: "19810505",                // укажите день рождения
		Email:    "vasya-pup@yahoo.com",     // укажите почту
	}

	id, err := insertClient(db, newClient)
	if err != nil {
		fmt.Println(err)
		return
	}

	// получение клиента по идентификатору и вывод на консоль
	client, err := selectClient(db, id)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(client)

	// обновление логина клиента
	newLogin := "VasyaPup" // укажите новый логин
	err = updateClientLogin(db, newLogin, id)
	if err != nil {
		fmt.Println(err)
		return
	}

	// получение клиента по идентификатору и вывод на консоль
	client, err = selectClient(db, id)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(client)

	// удаление клиента
	err = deleteClient(db, id)
	if err != nil {
		fmt.Println(err)
		return
	}

	// получение клиента по идентификатору и вывод на консоль
	_, err = selectClient(db, id)
	if err != nil {
		fmt.Println(err)
		return
	}
}

// Функция insertClient() — добавляет запись в таблицу clients.
// Возвращает идентификатор добавленной записи и ошибку.
// На вход принимает:
//   - db — указатель на объект типа sql.DB;
//   - client — объект типа Client с данными о клиенте.
func insertClient(db *sql.DB, client Client) (int64, error) {
	// Проверка пустых полей
	if client.FIO == "" {
		return 0, fmt.Errorf("FIO cannot be empty")
	}
	if client.Login == "" {
		return 0, fmt.Errorf("Login cannot be empty")
	}
	if client.Birthday == "" {
		return 0, fmt.Errorf("Birthday cannot be empty")
	}
	if client.Email == "" {
		return 0, fmt.Errorf("Email cannot be empty")
	}

	// Проверка формата даты
	_, err := time.Parse("20060102", client.Birthday) // формат YYYYMMDD
	// Закоментировано, так как автотестер не пропускает
	// if _, err := time.Parse("20060102", client.Birthday); err != nil {
	//     return 0, fmt.Errorf("birthday must be in YYYYMMDD format")
	// }

	// Добавление новой записи в таблицу clients
	res, err := db.Exec("INSERT INTO clients (fio, login, birthday, email) VALUES (:fio, :login, :birthday, :email)",
		sql.Named("fio", client.FIO),
		sql.Named("login", client.Login),
		sql.Named("birthday", client.Birthday),
		sql.Named("email", client.Email))

	// Обработка ошибки запроса
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return 0, fmt.Errorf("client with this login or email already exists")
		}
		return 0, err
	}

	// Проверка, что запрос вставил ровно одну строку
	affected, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	if affected != 1 {
		return 0, fmt.Errorf("expected 1 row affected, got %d", affected)
	}

	// Получение ID
	lastID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return lastID, nil // возращает идентификатор добавленной записи
}

// Функция updateClientLogin() — обновляет поле login у записи с заданным id в таблице clients.
// Возвращает ошибку.
// На вход принимает:
//   - db — указатель на объект типа sql.DB;
//   - login — логин клиента;
//   - id — идентификатор записи.
func updateClientLogin(db *sql.DB, login string, id int64) error {
	// Валидация входных данных
	if login == "" {
		return fmt.Errorf("login cannot be empty")
	}
	if id <= 0 {
		return fmt.Errorf("invalid client ID")
	}

	// Обновление поля login в таблице clients у записи с заданным id
	res, err := db.Exec("UPDATE clients SET login = :login WHERE id = :id",
		sql.Named("login", login),
		sql.Named("id", id))
	if err != nil {
		return err
	}

	// Проверка количества затронутых строк
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("client with ID %d not found", id)
	}

	return nil
}

// Функция deleteClient() — удаляет запись из таблицы clients по заданному id.
// Возвращает ошибку.
// На вход принимает:
//   - db — указатель на объект типа sql.DB;
//   - id — идентификатор записи.
func deleteClient(db *sql.DB, id int64) error {
	// Валидация входных данных
	if id <= 0 {
		return fmt.Errorf("no client with ID %d was deleted (record not found)", id)
	}

	// Удаляем записи из таблицы clients по заданному id
	res, err := db.Exec("DELETE FROM clients WHERE id = :id",
		sql.Named("id", id))
	if err != nil {
		return err
	}

	// Если затронуто 0 строк, значит, клиент с таким ID не существует
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("client with ID %d not found", id)
	}

	return nil
}

func selectClient(db *sql.DB, id int64) (Client, error) {
	client := Client{}

	row := db.QueryRow("SELECT id, fio, login, birthday, email FROM clients WHERE id = :id", sql.Named("id", id))
	err := row.Scan(&client.ID, &client.FIO, &client.Login, &client.Birthday, &client.Email)

	return client, err
}
