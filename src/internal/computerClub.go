package internal

import (
	"bufio"
	"fmt"
	"math"
	"slices"
	"time"
)

// События
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

// ComputerClub структура клуба
type ComputerClub struct {
	ClientsQueue []*Client          // Очередь клиентов
	TablesCount  int                // Общее количество столов
	Price        int                // Цена за час
	OpenTime     time.Time          // Время открытия
	CloseTime    time.Time          // Время закрытия
	Tables       map[int]*Table     // Столы и кто их занял
	Clients      map[string]*Client // Список клиентов
	FreeTables   int                // Количество свободных столов
}

// Client структура клиента
type Client struct {
	Name      string    // Имя клиента
	Table     int       // Стол который он занимает
	StartPlay time.Time // Время когда сел за стол
}

type Table struct {
	Client    *Client   // Кто занял стол
	Index     int       // Номер стола
	TotalTime time.Time // Общее время за столом
	Income    int       // Доход от стола
}

// NewClient создает нового клиента
func NewClient(name string) *Client {
	return &Client{
		Name: name,
	}
}

// NewComputerClub создает новый клуб
func NewComputerClub(totalTables, price int, start, end time.Time) *ComputerClub {
	return &ComputerClub{
		TablesCount:  totalTables,
		Price:        price,
		OpenTime:     start,
		CloseTime:    end,
		Tables:       NewTables(totalTables),
		ClientsQueue: make([]*Client, 0),
		Clients:      make(map[string]*Client),
		FreeTables:   totalTables,
	}
}

// NewTables создает map со столами
func NewTables(totalTable int) map[int]*Table {
	tables := map[int]*Table{}
	for i := 1; i <= totalTable; i++ {
		tables[i] = NewTable(i)
	}

	return tables
}

// NewTable создает новый стол
func NewTable(idx int) *Table {
	return &Table{
		TotalTime: time.Time{},
		Income:    0,
		Index:     idx,
	}
}

// StartWork начинает работу клуба
func (cc *ComputerClub) StartWork(scanner *bufio.Scanner) {
	fmt.Println(cc.OpenTime.Format("15:04"))

	for scanner.Scan() {
		line := scanner.Text()

		// Считываем по строчно файлик, чтобы не было переполнения
		eventTime, eventID, clientName, tableIdx, _, _ := GetAction(line)

		// Выводим входящее событие
		fmt.Println(line)

		// Человек пришел раньше времени открытия или позже закрытия
		if eventTime.Before(cc.OpenTime) || eventTime.After(cc.CloseTime) {
			fmt.Printf("%v %d %s\n",
				eventTime.Format("15:04"),
				CLIENT_ERROR,
				NOT_OPEN_YET)

			continue
		}

		switch eventID {
		// Человек пришел в клуб
		case CLIENT_ARRIVED:
			// Человек уже в клубе
			if _, ok := cc.Clients[clientName]; ok {
				fmt.Printf("%v %d %s\n",
					eventTime.Format("15:04"),
					CLIENT_ERROR,
					YOU_SHALL_NOT_PASS)
				continue
			}

			// Создаем нового клиента
			newClient := NewClient(clientName)

			// Добавляем человека в очередь ожидания и в список посетителей
			cc.ClientsQueue = append(cc.ClientsQueue, newClient)
			cc.Clients[clientName] = newClient

			// Событие, что клиент ждет
		case CLIENT_WAIT:
			// В ТЗ не написано, нужно ли смотреть на позицию человека в очереди
			// поэтому я решил просто убирать человека из очереди ожидания, если такой ивент случился

			// Если очередь больше, чем столов, то человек покидает клуб
			if len(cc.ClientsQueue) > cc.TablesCount {
				fmt.Printf("%v %d %s\n",
					eventTime.Format("15:04"),
					CLIENT_LEFT_OUTGOING,
					clientName)

				// Ищем позицию клиента в очереди ожидания
				clientInd := cc.findClientIndex(clientName)

				// удаляем клиента из очереди ожидания и списка посетителей
				cc.ClientsQueue = append(cc.ClientsQueue[:clientInd],
					cc.ClientsQueue[clientInd+1:]...)

				delete(cc.Clients, clientName)

				continue
			}

			// Если есть свободные столы, то клиент не может ждать
			if cc.FreeTables > 0 {
				fmt.Printf("%v %d %s\n",
					eventTime.Format("15:04"),
					CLIENT_ERROR,
					I_CAN_WAIT_NO_LONGER)
			}

		case CLIENT_LEFT:
			// Если клиент не находится в клубе
			if _, ok := cc.Clients[clientName]; !ok {
				fmt.Printf("%v %d %s\n",
					eventTime.Format("15:04"),
					CLIENT_ERROR,
					CLIENT_UNKNOWN)
				continue
			}

			// Освобождаем место и приглашаем следующего
			if len(cc.ClientsQueue) > 0 {
				nextClient := cc.ClientsQueue[0] // Берем клиента из очереди
				nextClient.StartPlay = eventTime // Ставим у следующего клиента начало время игры

				cc.ClientsQueue = cc.ClientsQueue[1:] // Обновляем очередь

				clientToLeave := cc.Clients[clientName] // Получаем клиента

				cc.Tables[clientToLeave.Table].Income += // Прибавляем доход
					cc.addIncome(eventTime, clientToLeave)

				// Получаем накопленное время стола
				totalTime := cc.Tables[clientToLeave.Table].TotalTime

				// Прибавляем к накопленному времени стола время, проведенное клиентом за столом
				cc.Tables[clientToLeave.Table].TotalTime =
					cc.addTime(eventTime, totalTime, clientToLeave)

				table := cc.Clients[clientName].Table // Получаем освободившейся стол
				nextClient.Table = table              // Сохраняем у клиента стол
				cc.Tables[table].Client = nextClient  // Клиент сел за стол

				delete(cc.Clients, clientName) // Удаляем старого клиента

				fmt.Printf("%v %d %s %d\n",
					eventTime.Format("15:04"),
					CLIENT_SAT_TABLE_OUTGOING,
					nextClient.Name,
					table)

				continue
			}

			// Просто освобождаем место у стола
			clientToLeave := cc.Clients[clientName]
			// Прибавляем доход
			cc.Tables[clientToLeave.Table].Income +=
				cc.addIncome(eventTime, clientToLeave)
			// Прибавляем время, проведенное клиентом за столом
			cc.Tables[clientToLeave.Table].TotalTime =
				cc.addTime(eventTime, cc.Tables[clientToLeave.Table].TotalTime, clientToLeave)
			// Убираем клиента из стола
			cc.Tables[clientToLeave.Table].Client = nil

			cc.FreeTables++ // Освободился стол

			delete(cc.Clients, clientName)

		case CLIENT_SAT_TABLE:
			// Если такого клиента нет в клубе
			if _, ok := cc.Clients[clientName]; !ok {
				fmt.Printf("%v %d %s\n",
					eventTime.Format("15:04"),
					CLIENT_ERROR,
					CLIENT_UNKNOWN)
				continue
			}

			currClient := cc.Clients[clientName]

			// Если стол пустой
			if t, ok := cc.Tables[tableIdx]; ok && t.Client == nil {
				currClient.StartPlay = eventTime            // Ставим время начала игры
				currClient.Table = tableIdx                 // Сохраняем у клиента стол
				cc.Tables[tableIdx].Client = currClient     // Посадили клиента за стол
				clientIdx := cc.findClientIndex(clientName) // Ищем позицию клиента в очереди

				cc.ClientsQueue = append(cc.ClientsQueue[:clientIdx],
					cc.ClientsQueue[clientIdx+1:]...) // Обновляем очередь

				cc.FreeTables-- // Стол занят

				continue
			}

			// Если стол занят
			if t, ok := cc.Tables[tableIdx]; ok && t.Client != nil {
				fmt.Printf("%v %d %s\n",
					eventTime.Format("15:04"),
					CLIENT_ERROR,
					PLACE_IS_BUSY)
			}
		}
	}

	// Создадим слайс для оставшихся (после закрытия клуба) клиентов
	// и посчитаем доходы с них и время за столом
	remainingClients := make([]string, 0, len(cc.Clients))
	for _, c := range cc.Clients {
		// Если клиент был за столом
		if c.Table != 0 {
			cc.Tables[c.Table].Income +=
				cc.addIncome(cc.CloseTime, c)

			cc.Tables[c.Table].TotalTime =
				cc.addTime(cc.CloseTime, cc.Tables[c.Table].TotalTime, c)
		}
		remainingClients = append(remainingClients, c.Name)
	}

	// Сортируем по имени клиентов
	slices.Sort(remainingClients)

	// Выводи оставшихся клиентов
	for _, name := range remainingClients {
		fmt.Printf("%v %d %s\n",
			cc.CloseTime.Format("15:04"),
			CLIENT_LEFT_OUTGOING,
			name)
	}

	// Время закрытия клуба
	fmt.Println(cc.CloseTime.Format("15:04"))

	// Выводим номер стола, заработок, общее время за столом
	for _, t := range cc.Tables {
		fmt.Printf("%d %d %v\n", t.Index, t.Income, t.TotalTime.Format("15:04"))
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

func (cc *ComputerClub) addTime(eventTime, totalTime time.Time,
	clientToLeave *Client) time.Time {
	t := totalTime.
		Add(eventTime.
			Sub(clientToLeave.StartPlay))
	return t
}

func (cc *ComputerClub) addIncome(eventTime time.Time,
	clientToLeave *Client) int {
	income := cc.Price * (int(math.Ceil(eventTime. // Округляем в большую сторону
							Sub(clientToLeave.StartPlay). // Вычитаем время ивента - начала игры
							Abs().
							Hours())))
	return income
}
