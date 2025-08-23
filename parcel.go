package main

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// Parcel описывает запись о посылке
type Parcel struct {
	Number    int
	Client    int
	Status    string
	Address   string
	CreatedAt string
}

// ParcelStore инкапсулирует работу с БД (таблица parcel)
type ParcelStore struct {
	db *sql.DB
}

func NewParcelStore(db *sql.DB) ParcelStore {
	return ParcelStore{db: db}
}

// Add добавляет запись в таблицу parcel и возвращает присвоенный номер посылки.
func (s ParcelStore) Add(p Parcel) (int, error) {
	res, err := s.db.Exec(`INSERT INTO parcel (client, status, address, created_at) VALUES (?,?,?,?)`,
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

// GetByNumber возвращает посылку по её номеру.
func (s ParcelStore) GetByNumber(number int) (Parcel, error) {
	var p Parcel
	row := s.db.QueryRow(`SELECT number, client, status, address, created_at FROM parcel WHERE number = ?`, number)
	err := row.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
	return p, err
}

// GetByClient возвращает список посылок по id клиента.
func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	rows, err := s.db.Query(`SELECT number, client, status, address, created_at FROM parcel WHERE client = ? ORDER BY number ASC`, client)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Parcel
	for rows.Next() {
		var p Parcel
		if err := rows.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// SetAddress меняет адрес, но только если статус = registered.
func (s ParcelStore) SetAddress(number int, address string) error {
	res, err := s.db.Exec(`UPDATE parcel SET address = ? WHERE number = ? AND status = ?`, address, number, ParcelStatusRegistered)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("cannot change address: not in registered status or parcel not found")
	}
	return nil
}

// SetStatus обновляет статус посылки.
func (s ParcelStore) SetStatus(number int, status string) error {
	res, err := s.db.Exec(`UPDATE parcel SET status = ? WHERE number = ?`, status, number)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// Delete удаляет запись, но только если статус = registered.
func (s ParcelStore) Delete(number int) error {
	res, err := s.db.Exec(`DELETE FROM parcel WHERE number = ? AND status = ?`, number, ParcelStatusRegistered)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("cannot delete: not in registered status or parcel not found")
	}
	return nil
}

// ParcelService содержит бизнес-логику и использует ParcelStore.
type ParcelService struct {
	store ParcelStore
}

func NewParcelService(store ParcelStore) ParcelService {
	return ParcelService{store: store}
}

// Register регистрирует новую посылку.
func (s ParcelService) Register(client int, address string) (Parcel, error) {
	p := Parcel{
		Client:    client,
		Status:    ParcelStatusRegistered,
		Address:   address,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	num, err := s.store.Add(p)
	if err != nil {
		return Parcel{}, err
	}
	p.Number = num
	fmt.Printf("Новая посылка Nº %d на адрес %s от клиента с идентификатором %d зарегистрирована %s\n", p.Number, p.Address, p.Client, p.CreatedAt)
	return p, nil
}

// Send устанавливает статус sent (только из registered).
func (s ParcelService) Send(number int) error {
	p, err := s.store.GetByNumber(number)
	if err != nil {
		return err
	}
	if p.Status != ParcelStatusRegistered {
		return errors.New("send allowed only from registered")
	}
	if err := s.store.SetStatus(number, ParcelStatusSent); err != nil {
		return err
	}
	fmt.Printf("у посылки Nº %d новый статус: %s\n", number, ParcelStatusSent)
	return nil
}

// Deliver устанавливает статус delivered (только из sent).
func (s ParcelService) Deliver(number int) error {
	p, err := s.store.GetByNumber(number)
	if err != nil {
		return err
	}
	if p.Status != ParcelStatusSent {
		return errors.New("deliver allowed only from sent")
	}
	if err := s.store.SetStatus(number, ParcelStatusDelivered); err != nil {
		return err
	}
	fmt.Printf("у посылки Nº %d новый статус: %s\n", number, ParcelStatusDelivered)
	return nil
}

// SetAddress меняет адрес (только пока registered).
func (s ParcelService) SetAddress(number int, address string) error {
	return s.store.SetAddress(number, address)
}

// Delete удаляет посылку (только пока registered).
func (s ParcelService) Delete(number int) error {
	return s.store.Delete(number)
}

// PrintClientParcels печатает список посылок клиента.
func (s ParcelService) PrintClientParcels(client int) error {
	fmt.Printf("Посылки клиента Nº %d:\n", client)
	parcels, err := s.store.GetByClient(client)
	if err != nil {
		return err
	}
	for _, p := range parcels {
		fmt.Printf("Посылка Nº %d на адрес %s от клиента с идентификатором %d зарегистрирована %s, статус %s\n",
			p.Number, p.Address, p.Client, p.CreatedAt, p.Status)
	}
	return nil
}
