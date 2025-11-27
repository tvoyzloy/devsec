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

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
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
			v, err := strconv.ParseInt(strings.TrimSpace(parts[i]), 10, 64)
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

		// data is valid → reset error counter
		errorCount = 0

		// 1) Load Average
		loadAvg := vals[0]
		if loadAvg > 30 {
			fmt.Printf("Load Average is too high: %d\n", loadAvg)
		}

		// 2) Memory usage
		totalMem := vals[1]
		usedMem := vals[2]
		if totalMem > 0 {
			percent := usedMem * 100 / totalMem
			if percent > 80 {
				fmt.Printf("Memory usage too high: %d%%\n", percent)
			}
		}

		// 3) Disk free space
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

		// 4) Network bandwidth
		totalNet := vals[5]
		usedNet := vals[6]
		if totalNet > 0 && usedNet > totalNet*9/10 {
			freeMbit := (totalNet - usedNet) * 8 / 1024 / 1024
			fmt.Printf("Network bandwidth usage high: %d Mbit/s available\n", freeMbit)
		}

		time.Sleep(time.Second)
	}
}
