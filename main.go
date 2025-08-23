package main

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

const (
	ParcelStatusRegistered = "registered"
	ParcelStatusSent       = "sent"
	ParcelStatusDelivered  = "delivered"
)

func main() {
	// Подключение к SQLite
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()

	store := NewParcelStore(db)
	service := NewParcelService(store)

	client := 1

	// 1) Регистрация первой посылки
	p1, err := service.Register(client, "Псков, д. Пушкина, ул. Колотушкина, д. 5")
	if err != nil {
		fmt.Println(err)
		return
	}

	// 2) Пока статус registered можно поменять адрес
	if err := service.SetAddress(p1.Number, "Саратов, д. Верхние Зори, ул. Козлова, д. 25"); err != nil {
		fmt.Println(err)
		return
	}

	// 3) Отправляем первую посылку
	if err := service.Send(p1.Number); err != nil {
		fmt.Println(err)
		return
	}

	// 4) Печать списка посылок клиента (должна отобразиться изменённая запись)
	if err := service.PrintClientParcels(client); err != nil {
		fmt.Println(err)
		return
	}

	// 5) Ещё раз печать списка (повтор для демонстрации)
	if err := service.PrintClientParcels(client); err != nil {
		fmt.Println(err)
		return
	}

	// 6) Регистрируем вторую посылку и сразу удаляем (разрешено только в registered)
	p2, err := service.Register(client, "Псков, д. Пушкина, ул. Колотушкина, д. 5")
	if err != nil {
		fmt.Println(err)
		return
	}

	if err := service.Delete(p2.Number); err != nil {
		fmt.Println(err)
		return
	}

	// 7) Финальная печать — второй посылки быть не должно
	if err := service.PrintClientParcels(client); err != nil {
		fmt.Println(err)
		return
	}
}
