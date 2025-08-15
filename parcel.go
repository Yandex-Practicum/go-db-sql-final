package main

import (
	"database/sql"
	"log"
)

type ParcelStore struct {
	db *sql.DB
}

func NewParcelStore(db *sql.DB) ParcelStore {
	return ParcelStore{db: db}
}

func (s ParcelStore) Add(p Parcel) (int, error) {

	// реализуйте добавление строки в таблицу parcel, используйте данные из переменной p
	res, err := s.db.Exec("INSERT INTO parcel (Client, Status, Address, Created_At) VALUES ( :Client, :Status, :Address, :CreatedAt)",
		sql.Named("Client", p.Client), sql.Named("Status", p.Status), sql.Named("Address", p.Address), sql.Named("CreatedAt", p.CreatedAt))

	if err != nil {
		return 0, err
	}
	// верните идентификатор последней добавленной записи
	id, err := res.LastInsertId()
	return int(id), err
}

func (s ParcelStore) Get(number int) (Parcel, error) {
	// реализуйте чтение строки по заданному number
	// здесь из таблицы должна вернуться только одна строка
	// заполните объект Parcel данными из таблицы
	p := Parcel{}

	err := s.db.QueryRow("SELECT Number, Client, Status, Address, Created_At FROM parcel WHERE Number = :Number",
		sql.Named("Number", number)).
		Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)

	if err != nil {
		return p, err
	}
	return p, err
}

func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	// реализуйте чтение строк из таблицы parcel по заданному client
	// здесь из таблицы может вернуться несколько строк
	rows, err := s.db.Query("SELECT Number, Client, Status, Address, Created_at FROM parcel WHERE Client = :Client",
		sql.Named("Client", client))
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	// заполните срез Parcel данными из таблицы
	var res []Parcel

	for rows.Next() {
		p := Parcel{}

		if err := rows.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt); err != nil {
			log.Fatal(err)
			return res, err
		}

		res = append(res, p)
	}

	return res, nil
}

func (s ParcelStore) SetStatus(number int, status string) error {
	// реализуйте обновление статуса в таблице parcel
	_, err := s.db.Exec("UPDATE parcel SET Status = :Status WHERE Number = :Number",
		sql.Named("Number", number), sql.Named("Status", status))

	return err
}

func (s ParcelStore) SetAddress(number int, address string) error {
	// реализуйте обновление адреса в таблице parcel
	// менять адрес можно только если значение статуса registered

	_, err := s.db.Exec("UPDATE parcel SET Address = :Address WHERE Number = :Number and Status = :Status",
		sql.Named("Number", number), sql.Named("Status", ParcelStatusRegistered), sql.Named("Address", address))

	return err
}

func (s ParcelStore) Delete(number int) error {
	// реализуйте удаление строки из таблицы parcel
	// удалять строку можно только если значение статуса registered
	_, err := s.db.Exec("DELETE FROM parcel WHERE Number = :Number and Status = :Status",
		sql.Named("Number", number), sql.Named("Status", ParcelStatusRegistered))

	return err
}
