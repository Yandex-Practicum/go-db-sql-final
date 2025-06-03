package main

import (
	"database/sql"
	"errors"
)

type ParcelStore struct {
	db *sql.DB
}

func NewParcelStore(db *sql.DB) ParcelStore {
	return ParcelStore{db: db}
}

func (s ParcelStore) Add(p Parcel) (int, error) {
	stmt := `INSERT INTO parcel (client, status, address, created_at) VALUES (?, ?, ?, ?)`
	res, err := s.db.Exec(stmt, p.Client, p.Status, p.Address, p.CreatedAt)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	return int(id), err
}

func (s ParcelStore) Get(number int) (Parcel, error) {
	stmt := `SELECT number, client, status, address, created_at FROM parcel WHERE number = ?`
	row := s.db.QueryRow(stmt, number)
	var p Parcel
	err := row.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Parcel{}, errors.New("посылка не найдена")
		}
		return Parcel{}, err
	}
	return p, nil
}

func (s ParcelStore) Delete(number int) error {
	p, err := s.Get(number)
	if err != nil {
		return err
	}
	if p.Status != ParcelStatusRegistered {
		return errors.New("нельзя удалить посылку, которая уже отправлена или доставлена")
	}
	stmt := `DELETE FROM parcel WHERE number = ?`
	_, err = s.db.Exec(stmt, number)
	return err
}

func (s ParcelStore) SetStatus(number int, status string) error {
	stmt := `UPDATE parcel SET status = ? WHERE number = ?`
	_, err := s.db.Exec(stmt, status, number)
	return err
}

func (s ParcelStore) SetAddress(number int, address string) error {
	stmt := `UPDATE parcel SET address = ? WHERE number = ?`
	_, err := s.db.Exec(stmt, address, number)
	return err
}

func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	stmt := `SELECT number, client, status, address, created_at FROM parcel WHERE client = ?`
	rows, err := s.db.Query(stmt, client)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var parcels []Parcel
	for rows.Next() {
		var p Parcel
		if err := rows.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt); err != nil {
			return nil, err
		}
		parcels = append(parcels, p)
	}
	return parcels, nil
}
