package main

import (
	"database/sql"
	"fmt"
)

type ParcelStore struct {
	db *sql.DB
}

func NewParcelStore(db *sql.DB) ParcelStore {
	return ParcelStore{db: db}
}

func (s ParcelStore) Add(p Parcel) (int, error) {
	// реализуйте добавление строки в таблицу parcel, используйте данные из переменной p
	res, err := s.db.Exec("INSERT INTO parcel (client, status, address, createdat) VALUES (:client, :status, :address, :createdat)",
		sql.Named("client", p.Client),
		sql.Named("status", p.Status),
		sql.Named("address", p.Address),
		sql.Named("createdat", p.CreatedAt))

	if err != nil {
		return 0, err
	}

	// верните идентификатор последней добавленной записи
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

func (s ParcelStore) Get(number int) (Parcel, error) {
	// реализуйте чтение строки по заданному number
	// здесь из таблицы должна вернуться только одна строка

	// заполните объект Parcel данными из таблицы
	p := Parcel{}
	row := s.db.QueryRow("SELECT number, client, status, address, createdat FROM parcel WHERE id = :id", sql.Named("id", number))
	err := row.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
	if err != nil {
		return p, err
	}

	return p, nil
}

func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	// реализуйте чтение строк из таблицы parcel по заданному client
	// здесь из таблицы может вернуться несколько строк
	var res []Parcel
	// заполните срез Parcel данными из таблицы

	rows, err := s.db.Query("SELECT number, client, status, address, createdat FROM parcel WHERE id = :id", sql.Named("id", client))
	if err != nil {
		return res, err
	}
	defer rows.Close()

	for rows.Next() {
		var p Parcel
		err := rows.Scan(&p.Number, &p.Client, &p.Address, &p.CreatedAt)
		if err != nil {
			return res, err
		}
		res = append(res, p)
	}
	return res, nil
}

func (s ParcelStore) SetStatus(number int, status string) error {
	// реализуйте обновление статуса в таблице parcel
	_, err := s.db.Exec("UPDATE parcel SET status = :status WHERE id = :id",
		sql.Named("status", status), sql.Named("id", number))
	if err != nil {
		return err
	}
	return nil
}

func (s ParcelStore) SetAddress(number int, address string) error {
	// реализуйте обновление адреса в таблице parcel
	// менять адрес можно только если значение статуса registered

	// Проверяем текущий статус посылки
	parcel, err := s.Get(number)
	if err != nil {
		return fmt.Errorf("ошибка получения посылки для проверки статуса: %w", err)
	}

	// Адрес можно изменить только для посылок со статусом 'registered'
	if parcel.Status != ParcelStatusRegistered {
		return fmt.Errorf("изменить адрес можно только для посылок со статусом 'registered'")
	}

	// Выполняем SQL-запрос на обновление адреса
	_, err = s.db.Exec("UPDATE parcel SET address = ? WHERE number = ?", address, number)
	if err != nil {
		return fmt.Errorf("ошибка обновления адреса посылки: %w", err)
	}
	return nil
}

func (s ParcelStore) Delete(number int) error {
	// реализуйте удаление строки из таблицы parcel
	// удалять строку можно только если значение статуса registered

	// Проверяем текущий статус посылки
	parcel, err := s.Get(number)
	if err != nil {
		return fmt.Errorf("ошибка получения посылки для проверки статуса: %w", err)
	}

	// Посылку можно удалить только со статусом 'registered'
	if parcel.Status != ParcelStatusRegistered {
		return fmt.Errorf("удалить посылку можно только со статусом 'registered'")
	}

	// Выполняем SQL-запрос на удаление записи
	_, err = s.db.Exec("DELETE FROM parcel WHERE number = ?", number)
	if err != nil {
		return fmt.Errorf("ошибка удаления посылки: %w", err)
	}
	return nil
}
