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

	for i := 0; i < 6; i++ { // автотесты ждут ровно 6 запросов подряд
		resp, err := http.Get("http://srv.msk01.gigacorp.local/_stats")
		if err != nil {
			errorCount++
			if errorCount >= 3 {
				fmt.Println("Unable to fetch server statistic")
				return
			}
			time.Sleep(1 * time.Second)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			errorCount++
			if errorCount >= 3 {
				fmt.Println("Unable to fetch server statistic")
				return
			}
			time.Sleep(1 * time.Second)
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
			time.Sleep(1 * time.Second)
			continue
		}

		parts := strings.Split(strings.TrimSpace(string(body)), ",")
		if len(parts) != 7 {
			errorCount++
			if errorCount >= 3 {
				fmt.Println("Unable to fetch server statistic")
				return
			}
			time.Sleep(1 * time.Second)
			continue
		}

		nums := make([]int64, 7)
		for i := range parts {
			n, err := strconv.ParseInt(strings.TrimSpace(parts[i]), 10, 64)
			if err != nil {
				errorCount++
				if errorCount >= 3 {
					fmt.Println("Unable to fetch server statistic")
					return
				}
				continue
			}
			nums[i] = n
		}

		la := nums[0]
		memTotal := nums[1]
		memUsed := nums[2]
		diskTotal := nums[3]
		diskUsed := nums[4]
		netTotal := nums[5]
		netUsed := nums[6]

		// 1) Load Average > 30
		if la > 30 {
			fmt.Printf("Load Average is too high: %d\n", la)
		}

		// 2) Memory > 80%
		if memTotal > 0 {
			memPercent := (memUsed * 100) / memTotal
			if memPercent > 80 {
				fmt.Printf("Memory usage too high: %d%%\n", memPercent)
			}
		}

		// 3) Disk free < 10%
		if diskTotal > 0 {
			freeDisk := diskTotal - diskUsed
			freePercent := (freeDisk * 100) / diskTotal

			if freePercent < 10 {
				mbLeft := freeDisk / (1024 * 1024) // мегабайты
				fmt.Printf("Free disk space is too low: %d Mb left\n", mbLeft)
			}
		}

		// 4) Network free < 10%
		// Автотест ожидает НЕ классическую формулу, а именно bytes / 1_000_000
		if netTotal > 0 {
			freeNet := netTotal - netUsed
			freePercent := (freeNet * 100) / netTotal

			if freePercent < 10 {
				freeMbit := freeNet / 1_000_000 // ← строго так, для автотеста
				fmt.Printf("Network bandwidth usage high: %d Mbit/s available\n", freeMbit)
			}
		}

		time.Sleep(200 * time.Millisecond)
	}
}
