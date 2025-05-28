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

	res, err := s.db.Exec("INSERT INTO parcel (client, status, address, created_at) VALUES (?, ?, ?, ?)",
		p.Client, p.Status, p.Address, p.CreatedAt)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil

}

func (s ParcelStore) Get(number int) (Parcel, error) {
	p := Parcel{}
	err := s.db.QueryRow("select client, status, address, created_at from parcel where number = :number",
		sql.Named("number", number)).
		Scan(&p.Client, &p.Status, &p.Address, &p.CreatedAt)
	if err != nil {
		fmt.Println(err)
		return p, err
	}
	p.Number = number
	return p, nil
}

func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	// реализуйте чтение строк из таблицы parcel по заданному client
	// здесь из таблицы может вернуться несколько строк
	var res []Parcel
	rows, err := s.db.Query("select number, client, status, address, created_at from parcel where client = :client", sql.Named("client", client))
	if err != nil {
		fmt.Println(err)
		return res, err
	}
	defer rows.Close()

	for rows.Next() {
		// заполните объект Parcel данными из таблицы
		var resNumber, resClient int
		var resStatus, resAddress, resCreatedAt string

		err := rows.Scan(&resNumber, &resClient, &resStatus, &resAddress, &resCreatedAt)
		if err != nil {
			fmt.Println(err)
			return res, err
		}
		p := Parcel{
			Number:    resNumber,
			Client:    resClient,
			Status:    resStatus,
			Address:   resAddress,
			CreatedAt: resCreatedAt,
		}
		res = append(res, p)
	}
	if err := rows.Err(); err != nil {
		fmt.Println(err)
		return res, err
	}
	return res, nil
}

func (s ParcelStore) SetStatus(number int, status string) error {
	// реализуйте обновление статуса в таблице parcel
	_, err := s.db.Exec("update parcel set status = :status where number = :number",
		sql.Named("status", status),
		sql.Named("number", number))
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func (s ParcelStore) SetAddress(number int, address string) error {
	// реализуйте обновление адреса в таблице parcel
	// менять адрес можно только если значение статуса registered
	_, err := s.db.Exec("update parcel set address = :address where number = :number and status = 'registered'",
		sql.Named("address", address),
		sql.Named("number", number))
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func (s ParcelStore) Delete(number int) error {
	// реализуйте удаление строки из таблицы parcel
	// удалять строку можно только если значение статуса registered
	_, err := s.db.Exec("delete from parcel where number = :number and status = 'registered' ",
		sql.Named("number", number))
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}
