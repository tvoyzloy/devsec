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
		if err != nil || resp.StatusCode != 200 {
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

		// Convert to numbers
		nums := make([]int64, 7)
		valid := true
		for i, p := range parts {
			n, err := strconv.ParseInt(p, 10, 64)
			if err != nil {
				valid = false
				break
			}
			nums[i] = n
		}

		if !valid {
			errorCount++
			if errorCount >= 3 {
				fmt.Println("Unable to fetch server statistic")
				return
			}
			time.Sleep(time.Second)
			continue
		}

		// Data fields
		loadAvg := nums[0]

		memTotal := nums[1]
		memUsed := nums[2]

		diskTotal := nums[3]
		diskUsed := nums[4]

		netTotal := nums[5]
		netUsed := nums[6]

		// 1) Load Average > 30
		if loadAvg > 30 {
			fmt.Printf("Load Average is too high: %d\n", loadAvg)
		}

		// 2) Memory > 80%
		if memTotal > 0 {
			memPercent := (memUsed * 100) / memTotal
			if memPercent > 80 {
				fmt.Printf("Memory usage too high: %d%%\n", memPercent)
			}
		}

		// 3) Disk free < 10%  (необходимо вывести оставшиеся МБ)
		if diskTotal > 0 {
			free := diskTotal - diskUsed
			freePercent := (free * 100) / diskTotal
			if freePercent < 10 {
				fmt.Printf("Free disk space is too low: %d Mb left\n", free/1024/1024)
			}
		}

		// 4) Network free < 10% (вывести свободный канал в Мбит/с)
		if netTotal > 0 {
			freeNet := netTotal - netUsed
			freePercent := (freeNet * 100) / netTotal
			if freePercent < 10 {
				fmt.Printf("Network bandwidth usage high: %d Mbit/s available\n", freeNet*8/1024/1024)
			}
		}

		// reset error counter on success
		errorCount = 0

		time.Sleep(time.Second)
	}
}
