package main

import (
	"database/sql"
	"log"
)

// ParcelStore implements the logic of working with storage
type ParcelStore struct {
	db *sql.DB
}

// NewParcelStore creates a new instance of the ParcelStore structure
func NewParcelStore(db *sql.DB) ParcelStore {
	return ParcelStore{db: db}
}

// Add adds a row to the parcel table
func (s ParcelStore) Add(p Parcel) (int, error) {
	// реализуйте добавление строки в таблицу parcel, используйте данные из переменной p
	res, err := s.db.Exec(`INSERT INTO parcel (client, status, address, created_at) 
	VALUES (:client, :status, :address, :created_at)`,
		sql.Named("client", p.Client),
		sql.Named("status", p.Status),
		sql.Named("address", p.Address),
		sql.Named("created_at", p.CreatedAt))

	if err != nil {
		log.Println(err)
		return 0, err
	}
	// верните идентификатор последней добавленной записи
	lastID, err := res.LastInsertId()
	if err != nil {
		log.Println(err)
		return 0, err
	}

	return int(lastID), nil
}

// Get subtracts a row from the parcel table by the specified number
func (s ParcelStore) Get(number int) (Parcel, error) {
	// реализуйте чтение строки по заданному number
	// здесь из таблицы должна вернуться только одна строка
	p := Parcel{}
	client := 0
	var status, address, createdAt string

	row := s.db.QueryRow("SELECT client, status, address, created_at FROM parcel WHERE number = :num",
		sql.Named("num", number))

	err := row.Scan(&client, &status, &address, &createdAt)
	if err != nil {
		log.Println(err)
		return p, err
	}
	// заполните объект Parcel данными из таблицы
	p = Parcel{
		Number:    number,
		Client:    client,
		Status:    status,
		Address:   address,
		CreatedAt: createdAt,
	}

	return p, nil
}

// GetByClient subtracts a row from the parcel table according to the specified client
func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	// реализуйте чтение строк из таблицы parcel по заданному client
	// здесь из таблицы может вернуться несколько строк
	var res []Parcel
	number := 0
	var status, address, createdAt string

	rows, err := s.db.Query(`SELECT number, status, address, created_at FROM parcel WHERE client = :client`,
		sql.Named("client", client))
	if err != nil {
		log.Println(err)
		return res, err
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&number, &status, &address, &createdAt)
		if err != nil {
			log.Println(err)
			return res, err
		}
		// заполните срез Parcel данными из таблицы
		p := Parcel{
			Number:    number,
			Client:    client,
			Status:    status,
			Address:   address,
			CreatedAt: createdAt,
		}

		res = append(res, p)
	}

	if err := rows.Err(); err != nil {
		log.Println(err)
		return res, err
	}

	return res, nil
}

// SetStatus updates the status in the parcel table
func (s ParcelStore) SetStatus(number int, status string) error {
	// реализуйте обновление статуса в таблице parcel
	_, err := s.db.Exec(`UPDATE parcel SET status = :status WHERE number = :number`,
		sql.Named("status", status),
		sql.Named("number", number))

	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

// SetAddress updates the address in the parcel table
func (s ParcelStore) SetAddress(number int, address string) error {
	// реализуйте обновление адреса в таблице parcel
	// менять адрес можно только если значение статуса registered
	_, err := s.db.Exec(`UPDATE parcel SET address = :address WHERE number = :number AND status = 'registered'`,
		sql.Named("address", address),
		sql.Named("number", number))

	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

// Delete deletes a row from the parcel table
func (s ParcelStore) Delete(number int) error {
	// реализуйте удаление строки из таблицы parcel
	// удалять строку можно только если значение статуса registered
	_, err := s.db.Exec("DELETE FROM parcel WHERE number = :number AND status = 'registered'",
		sql.Named("number", number))

	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}
