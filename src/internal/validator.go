package internal

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Доступные символы для имени клиента
const (
	validChars = "qwertyuiopasdfghjklzxcvbnm1234567890_"
)

// ValidateFile проверяет валидность файла
func ValidateFile(file *os.File) bool {
	scanner := bufio.NewScanner(file)

	var prevTime time.Time

	var flag bool

	// Получаем параметры для клуба и проверяем им корректность
	totalTables, _, _, _, invalidStr := GetOptions(scanner)
	if invalidStr != "" {
		fmt.Println(invalidStr)
		return false
	}

	// Проверяем следующие строки в файле на корректность
	for scanner.Scan() {
		nextTime, _, _, tableID, invalidStr, isValid := GetAction(scanner.Text())

		// Если это первое событие, то запоминаем время
		if !flag {
			if !isValid {
				fmt.Println(invalidStr)
				return false
			}

			prevTime = nextTime
			flag = true
			continue
		}

		// Если следующие событие оказалось раньше предыдущего
		// или его время совпадает с предыдущем,
		// или стол оказыватся > N || <= 0
		if !isValid || (nextTime.Before(prevTime) || nextTime.Equal(prevTime)) ||
			tableID > totalTables {

			fmt.Println(invalidStr)
			return false
		}

		prevTime = nextTime
	}

	// Если файл пустой, то вернется false
	return flag
}

// GetOptions получает параметры для клуба
// Проверяем первые 3 параметра
func GetOptions(scanner *bufio.Scanner) (int, int, time.Time, time.Time, string) {
	var tablesCount, price int
	var start, end time.Time
	var err error
	var line string
	for i := 0; i < 3; i++ {
		// Если было задано меньше 3 параметров, то вернем последнюю считанную строку
		if !scanner.Scan() {
			return 0, 0, time.Time{}, time.Time{}, line
		}
		line = scanner.Text()
		switch i {
		case 0:
			// Проверяем количество столов в клубе
			tablesCount, err = strconv.Atoi(line)
			if err != nil || tablesCount <= 0 {
				return 0, 0, time.Time{}, time.Time{}, line
			}
		case 1:
			// Проверяем начало и конец рабочего времени
			workDuration := strings.Split(line, " ")
			start, err = time.Parse("15:04", workDuration[0])
			if err != nil || len(workDuration) != 2 {
				return 0, 0, time.Time{}, time.Time{}, line
			}
			end, err = time.Parse("15:04", workDuration[1])
			if err != nil {
				return 0, 0, time.Time{}, time.Time{}, line
			}
		case 2:
			// Проверяем цену
			price, err = strconv.Atoi(line)
			if err != nil || price <= 0 {
				return 0, 0, time.Time{}, time.Time{}, line
			}
		}
	}

	return tablesCount, price, start, end, ""
}

func GetAction(line string) (time.Time, int, string, int, string, bool) {
	data := strings.Split(line, " ")

	if len(data) < 3 {
		return time.Time{}, 0, "", 0, line, false
	}

	eventTime, err := time.Parse("15:04", data[0])
	if err != nil {
		return time.Time{}, 0, "", 0, line, false
	}

	eventID, err := strconv.Atoi(data[1])
	if err != nil {
		return time.Time{}, 0, "", 0, line, false
	}

	if !clientIsValid(data[2]) {
		return time.Time{}, 0, "", 0, line, false
	}

	client := data[2]

	if len(data) == 3 {
		return eventTime, int(eventID), client, 0, line, true
	}

	tableID, err := strconv.Atoi(data[3])
	if err != nil || tableID <= 0 {
		return time.Time{}, 0, "", 0, line, false
	}

	return eventTime, int(eventID), client, tableID, line, true
}

func clientIsValid(client string) bool {
	for _, char := range client {
		if !strings.ContainsRune(validChars, char) {
			return false
		}
	}

	return true
}
