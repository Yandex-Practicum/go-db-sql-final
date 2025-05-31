package parcel

import (
	"database/sql"
	"fmt"
)

type ParcelStore struct {
	db *sql.DB
}

const (
	ParcelStatusRegistered = "registered"
	ParcelStatusSent       = "sent"
	ParcelStatusDelivered  = "delivered"
)

type Parcel struct {
	Number    int
	Client    int
	Status    string
	Address   string
	CreatedAt string
}

func NewParcelStore(db *sql.DB) ParcelStore {
	return ParcelStore{db: db}
}

func (s ParcelStore) Add(p Parcel) (int, error) {
	// реализуйте добавление строки в таблицу parcel, используйте данные из переменной p
	res, err := s.db.Exec("INSERT INTO parcel (number, client, status, address, createdAt) VALUES (:number, :client, :status, :address, :createdAt)",
		sql.Named("number", p.Number),
		sql.Named("client", p.Client),
		sql.Named("status", p.Status),
		sql.Named("address", p.Address),
		sql.Named("createdAt", p.CreatedAt))
	if err != nil {
		return 0, fmt.Errorf("error adding parcel: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("error getting ID: %w", err)
	}
	// верните идентификатор последней добавленной записи
	return int(id), nil
}

func (s ParcelStore) Get(number int) (Parcel, error) {
	// реализуйте чтение строки по заданному number
	// здесь из таблицы должна вернуться только одна строка
	var p Parcel
	query := "SELECT number, client, status, address FROM parcel WHERE number = ?"

	err := s.db.QueryRow(query, number).Scan(
		&p.Number,
		&p.Client,
		&p.Status,
		&p.Address,
		&p.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return Parcel{}, fmt.Errorf("parcel with number %d not found", number)
		}
		return Parcel{}, fmt.Errorf("error receiving parcel: %w", err)
	}

	return p, nil
}

func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	// реализуйте чтение строк из таблицы parcel по заданному client
	// здесь из таблицы может вернуться несколько строк

	// заполните срез Parcel данными из таблицы
	var res []Parcel
	query := "SELECT number, client, status, address, createdAt FROM parcel WHERE client = ?"

	rows, err := s.db.Query(query, client)
	if err != nil {
		return nil, fmt.Errorf("error receiving parcels: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var p Parcel
		err := rows.Scan(
			&p.Number,
			&p.Client,
			&p.Status,
			&p.Address,
			&p.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error with reading data: %w", err)
		}
		res = append(res, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error with reading data: %w", err)
	}

	return res, nil
}

func (s ParcelStore) SetStatus(number int, status string) error {
	// реализуйте обновление статуса в таблице parcel
	allowedStatuses := map[string]struct{}{
		ParcelStatusRegistered: {},
		ParcelStatusDelivered:  {},
		ParcelStatusSent:       {},
	}

	if _, ok := allowedStatuses[status]; !ok {
		return fmt.Errorf("invalid status: %s", status)
	}
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM parcel WHERE number = ?)"
	err := s.db.QueryRow(query, number).Scan(&exists)

	if err != nil {
		return fmt.Errorf("error checking existence of package: %w", err)
	}
	if !exists {
		return fmt.Errorf("parcel with number %d not found", number)
	}

	updateQuery := "UPDATE parcel SET status = ? WHERE number = ?"
	_, err = s.db.Exec(updateQuery, status, number)
	if err != nil {
		return fmt.Errorf("error with update status: %w", err)
	}

	return nil
}

func (s ParcelStore) SetAddress(number int, address string) error {
	// реализуйте обновление адреса в таблице parcel
	// менять адрес можно только если значение статуса registered
	var currentStatus string
	query := "SELECT status FROM parcel WHERE number = ?"
	err := s.db.QueryRow(query, number).Scan(&currentStatus)

	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("parcel with number %d not found", number)
		}
		return fmt.Errorf("error while receiving parcel status: %w", err)
	}
	if currentStatus != "registered" {
		return fmt.Errorf("changing the address is only possible for parcels with the status 'registered'")
	}

	_, err = s.db.Exec("UPDATE parcel SET address = :address WHERE number = :number",
		sql.Named("address", address),
		sql.Named("number", number))
	if err != nil {
		return fmt.Errorf("error with update address: %w", err)
	}
	return nil
}

func (s ParcelStore) Delete(number int) error {
	// реализуйте удаление строки из таблицы parcel
	// удалять строку можно только если значение статуса registered
	var currentStatus string

	query := "SELECT status FROM parcel WHERE number = ?"

	err := s.db.QueryRow(query, number).Scan(&currentStatus)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("parcel with number %d not found", number)
		}
		return fmt.Errorf("error while receiving parcel status: %w", err)
	}

	if currentStatus != "registered" {
		return fmt.Errorf("deletion is only possible for parcels with the status 'registered'")
	}

	deleteQuery := "DELETE FROM parcel WHERE number = ?"
	_, err = s.db.Exec(deleteQuery, number)
	if err != nil {
		return fmt.Errorf("error with deletion parcel: %w", err)
	}
	return nil
}
