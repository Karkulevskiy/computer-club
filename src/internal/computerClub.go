package internal

import (
	"bufio"
	"fmt"
	"time"
)

const (
	CLIENT_ARRIVED            = 1
	CLIENT_SAT_TABLE          = 2
	CLIENT_WAIT               = 3
	CLIENT_LEFT               = 4
	CLIENT_LEFT_OUTGOING      = 11
	CLIENT_SAT_TABLE_OUTGOING = 12
	CLIENT_ERROR              = 13
	YOU_SHALL_NOT_PASS        = "YouShallNotPass"
	NOT_OPEN_YET              = "NotOpenYet"
	PLACE_IS_BUSY             = "PlaceIsBusy"
	CLIENT_UNKNOWN            = "ClientUnknown"
	I_CAN_WAIT_NO_LONGER      = "ICanWaitNoLonger"
)

type ComputerClub struct {
	ClientsQueue []*Client
	TablesCount  int
	Price        int
	OpenTime     time.Time
	CloseTime    time.Time
	Tables       map[int]*Client
	Clients      map[string]*Client
	Income       int
}

type Client struct {
	Name      string
	Table     int
	StartPlay time.Time
}

func NewComputerClub(totalTables, price int, start, end time.Time) *ComputerClub {
	return &ComputerClub{
		TablesCount:  totalTables,
		Price:        price,
		OpenTime:     start,
		CloseTime:    end,
		Tables:       make(map[int]*Client),
		ClientsQueue: make([]*Client, 0),
		Clients:      make(map[string]*Client),
	}
}

func (cc *ComputerClub) StartWork(scanner *bufio.Scanner) {
	fmt.Println(cc.OpenTime)

	for scanner.Scan() {
		line := scanner.Text()

		eventTime, eventID, clientName, tableID, _ := GetAction(line)

		fmt.Println(line)

		// Человек пришел раньше времени открытия
		if eventTime.Before(cc.OpenTime) {
			fmt.Printf("%v %d %s\n", eventTime, CLIENT_ERROR, NOT_OPEN_YET)
			continue
		}

		switch eventID {
		// Человек уже есть в клубе, но событие, что человек пришел
		case CLIENT_ARRIVED:
			// Человек уже в клубе
			if _, ok := cc.Clients[clientName]; ok {
				fmt.Printf("%v %d %s\n", eventTime, CLIENT_ERROR, YOU_SHALL_NOT_PASS)
				continue
			}

			newClient := NewClient(clientName)
			// Добавляем человека в очередь ожидания
			cc.ClientsQueue = append(cc.ClientsQueue, newClient)
			cc.Clients[clientName] = newClient

		case CLIENT_WAIT:
			// В ТЗ не написано, нужно ли смотреть на позицию человека в очереди
			if len(cc.ClientsQueue) > cc.TablesCount {
				fmt.Printf("%v %d %s\n", eventTime, CLIENT_LEFT_OUTGOING, clientName)

				clientInd := cc.findClientIndex(clientName)

				// удаляем клиента из очереди ожидания
				cc.ClientsQueue = append(cc.ClientsQueue[:clientInd], cc.ClientsQueue[clientInd+1:]...)
				delete(cc.Clients, clientName)
				continue
			}
			fmt.Printf("%v %d %s\n", eventTime, CLIENT_ERROR, I_CAN_WAIT_NO_LONGER)

		case CLIENT_LEFT:
			// Если клиент не находится в клубе
			if _, ok := cc.Clients[clientName]; !ok {
				fmt.Printf("%v %d %s\n", eventTime, CLIENT_ERROR, CLIENT_UNKNOWN)
				continue
			}
			// Освобождаем место и приглашаем следующего
			if len(cc.ClientsQueue) > 0 {
				nextClient := cc.ClientsQueue[0]      // Берем клиента из очереди
				cc.ClientsQueue = cc.ClientsQueue[1:] // Обновляем очередь
				delete(cc.Clients, clientName)        // Удаляем старого клиента

				table := cc.Clients[clientName].Table    // Получаем освободившейся стол
				cc.Tables[table] = nextClient            // Клиент сел за стол
				cc.Clients[nextClient.Name] = nextClient // Закрепляем за клиентом его стол
			}
		}
	}
}

// findClientIndex находит позицию клиента в очереди
func (cc *ComputerClub) findClientIndex(clientName string) int {
	for i, c := range cc.ClientsQueue {
		if c.Name == clientName {
			return i
		}
	}
	return -1
}

func NewClient(name string) *Client {
	return &Client{
		Name: name,
	}
}
