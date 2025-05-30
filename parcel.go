package main

import (
	"database/sql"
	"errors"
	"fmt"
)

var (
	// ErrNotRegistered возвращается, если пытаются изменить адрес или удалить посылку,
	// когда её статус не registered
	ErrNotRegistered = errors.New("parcel status is not registered")
)

// ParcelStore хранит ссылку на БД
type ParcelStore struct {
	db *sql.DB
}

func NewParcelStore(db *sql.DB) ParcelStore {
	return ParcelStore{db: db}
}

// Add вставляет новую посылку и возвращает её сгенерированный номер (автоинкремент)
func (s ParcelStore) Add(p Parcel) (int, error) {
	res, err := s.db.Exec(
		`INSERT INTO parcel(client, status, address, created_at)
         VALUES(?, ?, ?, ?)`,
		p.Client, p.Status, p.Address, p.CreatedAt,
	)
	if err != nil {
		return 0, fmt.Errorf("Add: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("Add – LastInsertId: %w", err)
	}
	return int(id), nil
}

// Get читает одну посылку по её номеру
func (s ParcelStore) Get(number int) (Parcel, error) {
	var p Parcel
	row := s.db.QueryRow(
		`SELECT number, client, status, address, created_at
         FROM parcel
         WHERE number = ?`,
		number,
	)
	err := row.Scan(
		&p.Number,
		&p.Client,
		&p.Status,
		&p.Address,
		&p.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return p, fmt.Errorf("Get: parcel %d not found", number)
		}
		return p, fmt.Errorf("Get: %w", err)
	}
	return p, nil
}

// GetByClient возвращает все посылки данного клиента
func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	rows, err := s.db.Query(
		`SELECT number, client, status, address, created_at
         FROM parcel
         WHERE client = ?`,
		client,
	)
	if err != nil {
		return nil, fmt.Errorf("GetByClient: %w", err)
	}
	defer rows.Close()

	var res []Parcel
	for rows.Next() {
		var p Parcel
		if err := rows.Scan(
			&p.Number,
			&p.Client,
			&p.Status,
			&p.Address,
			&p.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("GetByClient scan: %w", err)
		}
		res = append(res, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GetByClient rows: %w", err)
	}
	return res, nil
}

// SetStatus меняет статус посылки
func (s ParcelStore) SetStatus(number int, status string) error {
	res, err := s.db.Exec(
		`UPDATE parcel
         SET status = ?
         WHERE number = ?`,
		status, number,
	)
	if err != nil {
		return fmt.Errorf("SetStatus: %w", err)
	}
	// Если нужно, можно проверить RowsAffected, но в NextStatus
	// мы не требуем ошибки, если посылка уже в финальном статусе.
	_ = res
	return nil
}

// SetAddress меняет адрес только если статус registered
func (s ParcelStore) SetAddress(number int, address string) error {
	res, err := s.db.Exec(
		`UPDATE parcel
         SET address = ?
         WHERE number = ? AND status = ?`,
		address, number, ParcelStatusRegistered,
	)
	if err != nil {
		return fmt.Errorf("SetAddress: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("SetAddress RowsAffected: %w", err)
	}
	if n == 0 {
		return ErrNotRegistered
	}
	return nil
}

// Delete удаляет посылку только если статус registered
func (s ParcelStore) Delete(number int) error {
	res, err := s.db.Exec(
		`DELETE FROM parcel
         WHERE number = ? AND status = ?`,
		number, ParcelStatusRegistered,
	)
	if err != nil {
		return fmt.Errorf("Delete: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("Delete RowsAffected: %w", err)
	}
	if n == 0 {
		return ErrNotRegistered
	}
	return nil
}
