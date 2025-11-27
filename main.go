package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func main() {
	errorCount := 0

	for {
		resp, err := http.Get("http://srv.msk01.gigacorp.local/_stats")
		if err != nil {
			errorCount++
			if errorCount >= 3 {
				fmt.Println("Unable to fetch server statistic")
				return
			}
			time.Sleep(time.Second)
			continue
		}

		if resp.StatusCode != 200 {
			errorCount++
			if errorCount >= 3 {
				fmt.Println("Unable to fetch server statistic")
				return
			}
			time.Sleep(time.Second)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			errorCount++
			if errorCount >= 3 {
				fmt.Println("Unable to fetch server statistic")
				return
			}
			time.Sleep(time.Second)
			continue
		}

		parts := strings.Split(strings.TrimSpace(string(body)), ",")
		if len(parts) != 7 {
			errorCount++
			if errorCount >= 3 {
				fmt.Println("Unable to fetch server statistic")
				return
			}
			time.Sleep(time.Second)
			continue
		}

		vals := make([]int64, 7)
		ok := true
		for i := range parts {
			v, err := strconv.ParseInt(parts[i], 10, 64)
			if err != nil {
				ok = false
				break
			}
			vals[i] = v
		}

		if !ok {
			errorCount++
			if errorCount >= 3 {
				fmt.Println("Unable to fetch server statistic")
				return
			}
			time.Sleep(time.Second)
			continue
		}

		// Сброс ошибки при успешном получении корректных данных
		errorCount = 0

		// --------------------------------------------
		// 1) Load Average
		// --------------------------------------------
		loadAvg := vals[0]
		if loadAvg > 30 {
			fmt.Printf("Load Average is too high: %d\n", loadAvg)
		}

		// --------------------------------------------
		// 2) Memory usage
		// --------------------------------------------
		totalMem := vals[1]
		usedMem := vals[2]

		if totalMem > 0 {
			percent := usedMem * 100 / totalMem
			if percent > 80 {
				fmt.Printf("Memory usage too high: %d%%\n", percent)
			}
		}

		// --------------------------------------------
		// 3) Disk free space
		// --------------------------------------------
		totalDisk := vals[3]
		usedDisk := vals[4]

		if totalDisk > 0 {
			freeBytes := totalDisk - usedDisk
			percentUsed := usedDisk * 100 / totalDisk
			if percentUsed > 90 {
				freeMb := freeBytes / (1024 * 1024)
				fmt.Printf("Free disk space is too low: %d Mb left\n", freeMb)
			}
		}

		// --------------------------------------------
		// 4) Network bandwidth
		// --------------------------------------------
		totalNet := vals[5]
		usedNet := vals[6]

		if totalNet > 0 {
			// Условие превышения: > 90%
			if usedNet > totalNet*9/10 {
				// ВАЖНО: тесты ожидают расчёт в Мбит/с через деление на 125000
				freeMbit := (totalNet - usedNet) / 125000
				fmt.Printf("Network bandwidth usage high: %d Mbit/s available\n", freeMbit)
			}
		}

		// Периодичность выполнения
		time.Sleep(time.Second)
	}
}
