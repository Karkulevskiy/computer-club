package internal

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	validChars = "qwertyuiopasdfghjklzxcvbnm1234567890_"
)

func ValidateFile(file *os.File) bool {
	scanner := bufio.NewScanner(file)

	if _, _, _, _, invalidStr := GetOptions(scanner); invalidStr != "" {
		fmt.Println(invalidStr)
		return false
	}

	for scanner.Scan() {
		if _, _, _, _, invalidStr := GetAction(scanner.Text()); invalidStr != "" {
			fmt.Println(invalidStr)
			return false
		}
	}

	return true
}

func GetOptions(scanner *bufio.Scanner) (int, int, time.Time, time.Time, string) {
	var tablesCount, price int
	var start, end time.Time
	var err error
	for i := 0; i < 3; i++ {
		scanner.Scan()
		line := scanner.Text()
		switch i {
		case 0:
			tablesCount, err = strconv.Atoi(line)
			if err != nil || tablesCount <= 0 {
				return 0, 0, time.Time{}, time.Time{}, line
			}
		case 1:
			workDuration := strings.Split(line, " ")
			start, err = time.Parse("15:04", workDuration[0])
			if err != nil {
				return 0, 0, time.Time{}, time.Time{}, line
			}
			end, err = time.Parse("15:04", workDuration[1])
			if err != nil {
				return 0, 0, time.Time{}, time.Time{}, line
			}
		case 2:
			price, err = strconv.Atoi(line)
			if err != nil || price <= 0 {
				return 0, 0, time.Time{}, time.Time{}, line
			}
		}
	}

	return tablesCount, price, start, end, ""
}

func GetAction(line string) (time.Time, int, string, int, string) {
	data := strings.Split(line, " ")

	if len(data) < 3 {
		return time.Time{}, 0, "", 0, line
	}

	eventTime, err := time.Parse("15:04", data[0])
	if err != nil {
		return time.Time{}, 0, "", 0, line
	}

	eventID, err := strconv.Atoi(data[1])
	if err != nil {
		return time.Time{}, 0, "", 0, line
	}

	if !clientIsValid(data[2]) {
		return time.Time{}, 0, "", 0, line
	}

	client := data[2]

	if len(data) == 3 {
		return eventTime, int(eventID), client, 0, ""
	}

	tableID, err := strconv.Atoi(data[3])
	if err != nil {
		return time.Time{}, 0, "", 0, line
	}

	return eventTime, int(eventID), client, tableID, ""
}

func clientIsValid(client string) bool {
	for _, char := range client {
		if !strings.ContainsRune(validChars, char) {
			return false
		}
	}

	return true
}
