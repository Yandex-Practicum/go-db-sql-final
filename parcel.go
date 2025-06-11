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
	res, err := s.db.Exec("insert into parcel (client,status,address,created_at) values (:client,:status,:address,:created_at)",
		sql.Named("client", p.Client),
		sql.Named("address", p.Address),
		sql.Named("status", p.Status),
		sql.Named("created_at", p.CreatedAt))
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
	str := s.db.QueryRow("select number,status,client,address,created_at from parcel where number=:number", sql.Named("number", number))
	// заполните объект Parcel данными из таблицы
	p := Parcel{}

	err := str.Scan(&p.Number, &p.Status, &p.Client, &p.Address, &p.CreatedAt)

	if err != nil {
		return p, err
	}

	return p, nil
}

func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	// реализуйте чтение строк из таблицы parcel по заданному client
	// здесь из таблицы может вернуться несколько строк
	str, err := s.db.Query("select number,status,client,address,created_at from parcel where client=:client", sql.Named("client", client))
	if err != nil {
		return nil, err
	}
	// заполните срез Parcel данными из таблицы
	var res []Parcel
	var item Parcel
	for str.Next() {
		err := str.Scan(&item.Number, &item.Status, &item.Client, &item.Address, &item.CreatedAt)
		if err != nil {
			return nil, err
		}
		res = append(res, item)
	}
	return res, nil
}

func (s ParcelStore) SetStatus(number int, status string) error {
	// реализуйте обновление статуса в таблице parcel

	if status == ParcelStatusDelivered || status == ParcelStatusRegistered || status == ParcelStatusSent {
		_, err := s.db.Exec("update parcel set status=:status where number=:number",
			sql.Named("status", status),
			sql.Named("number", number))

		if err != nil {
			return err
		}
		return nil
	}

	return fmt.Errorf("unknown status")
}

func (s ParcelStore) SetAddress(number int, address string) error {
	// реализуйте обновление адреса в таблице parcel
	// менять адрес можно только если значение статуса registered

	p, err := s.Get(number)
	if err != nil {
		return err
	}

	if p.Status == ParcelStatusRegistered {
		_, err := s.db.Exec("update parcel set address=:address where number=:number",
			sql.Named("address", address),
			sql.Named("number", number))
		if err != nil {
			return err
		}
	}

	return nil
}

func (s ParcelStore) Delete(number int) error {
	// реализуйте удаление строки из таблицы parcel
	// удалять строку можно только если значение статуса registered
	p, err := s.Get(number)
	if err != nil {
		return err
	}

	if p.Status == ParcelStatusRegistered {
		_, err := s.db.Exec("delete from parcel where number=:number",
			sql.Named("number", number))
		if err != nil {
			return err
		}
	}
	return nil
}
