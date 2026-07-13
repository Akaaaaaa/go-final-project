package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const DateFormat = "20060102"

func NextDate(now time.Time, dateStr string, repeat string) (string, error) {
	date, err := time.Parse(DateFormat, dateStr)
	if err != nil {
		return "", fmt.Errorf("неверный формат даты: %v", err)
	}
	if repeat == "" {
		return "", nil
	}

	parts := strings.Fields(repeat)
	if len(parts) == 0 {
		return "", fmt.Errorf("неверный формат правила повторения: %v", err)
	}

	rule := parts[0]

	switch rule {
	case "d":
		if len(parts) != 2 {
			return "", fmt.Errorf("для правила 'd' требуется указать количество дней: %v", err)
		}
		days, err := strconv.Atoi(parts[1])
		if err != nil || days <= 0 || days > 400 {
			return "", fmt.Errorf("неверное количество дней: %v", err)
		}
		for {
			date = date.AddDate(0, 0, days)
			if date.After(now) {
				return date.Format(DateFormat), nil
			}
		}
	case "y":
		if len(parts) != 1 {
			return "", fmt.Errorf("для правила 'y' не требуется дополнительных параметров: %v", err)
		}
		for {
			date = date.AddDate(1, 0, 0)
			if date.After(now) {
				return date.Format(DateFormat), nil
			}
		}
	case "w":
		if len(parts) != 2 {
			return "", fmt.Errorf("для правила 'w' требуется указать количество недель: %v", err)
		}

		weekDays := strings.Split(parts[1], ",")
		validDays := make(map[int]bool)
		for _, d := range weekDays {
			dayNum, err := strconv.Atoi(d)
			if err != nil || dayNum < 1 || dayNum > 7 {
				return "", fmt.Errorf("недопустимый день недели: %v", err)
			}
			validDays[dayNum] = true
		}
		for {
			date = date.AddDate(0, 0, 1)
			weekDayGo := int(date.Weekday())
			if weekDayGo == 0 {
				weekDayGo = 7
			}
			if date.After(now) && validDays[weekDayGo] {
				return date.Format(DateFormat), nil
			}
		}
	case "m":
		var daysStr, monthStr string
		if len(parts) == 2 {
			daysStr = parts[1]
			monthStr = ""
		} else if len(parts) == 3 {
			daysStr = parts[1]
			monthStr = parts[2]
		} else {
			return "", fmt.Errorf("неверный формат правила")
		}
		validDays := make(map[int]bool)
		dayParts := strings.Split(daysStr, ",")
		for _, d := range dayParts {
			dayNum, err := strconv.Atoi(d)
			if err != nil {
				return "", fmt.Errorf("неверный день месяца: %v", err)
			}
			if dayNum < -2 || dayNum == 0 || dayNum > 31 {
				return "", fmt.Errorf("число вышло за пределы")
			}
			validDays[dayNum] = true
		}
		validMonth := make(map[int]bool)
		if monthStr == "" {
			for i := 1; i <= 12; i++ {
				validMonth[i] = true
			}
		} else {
			monthParts := strings.Split(monthStr, ",")
			for _, m := range monthParts {
				monthNum, err := strconv.Atoi(m)
				if err != nil || monthNum < 1 || monthNum > 12 {
					return "", fmt.Errorf("неверный номер месяца: %s", m)
				}
				validMonth[monthNum] = true
			}
		}
		for {
			date = date.AddDate(0, 0, 1)
			if !date.After(now) {
				continue
			}

			if !validMonth[int(date.Month())] {
				continue
			}

			dayOfMonth := date.Day()
			lastDay := time.Date(date.Year(), date.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()
			preLastDay := lastDay - 1

			isValidDay := false
			if validDays[-1] && dayOfMonth == lastDay {
				isValidDay = true
			}
			if validDays[-2] && dayOfMonth == preLastDay {
				isValidDay = true
			}
			if validDays[dayOfMonth] {
				isValidDay = true
			}
			if isValidDay {
				return date.Format(DateFormat), nil
			}
		}
	default:
		return "", fmt.Errorf("неподдерживаемый формат правила: %s", rule)
	}
}

func NextDateHandler(w http.ResponseWriter, r *http.Request) {
	nowStr := r.FormValue("now")
	dateStr := r.FormValue("date")
	repeat := r.FormValue("repeat")

	var now time.Time
	var err error

	if nowStr == "" {
		now = time.Now()
	} else {
		now, err = time.Parse(DateFormat, nowStr)
		if err != nil {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Неверный формат даты 'now'. Ожидается YYYYMMDD"})
			return
		}
	}

	nextDate, err := NextDate(now, dateStr, repeat)
	if err != nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(nextDate))
}
