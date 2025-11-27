package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

const statsURL = "http://srv.msk01.gigacorp.local/_stats"

func main() {
	errorCount := 0

	for {
		values, ok := fetchStats()
		if !ok {
			errorCount = handleError(errorCount)
			continue
		}

		errorCount = 0
		report(values)
	}
}

func fetchStats() ([7]int64, bool) {
	var values [7]int64

	resp, err := http.Get(statsURL)
	if err != nil {
		return values, false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return values, false
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return values, false
	}

	parts := strings.Split(strings.TrimSpace(string(body)), ",")
	if len(parts) != len(values) {
		return values, false
	}

	for i, part := range parts {
		num, err := strconv.ParseInt(strings.TrimSpace(part), 10, 64)
		if err != nil {
			return values, false
		}
		values[i] = num
	}

	return values, true
}

func handleError(count int) int {
	count++
	if count >= 3 {
		fmt.Println("Unable to fetch server statistic")
		return 0
	}
	return count
}

func report(vals [7]int64) {
	loadAvg := vals[0]
	totalMem := vals[1]
	usedMem := vals[2]
	totalDisk := vals[3]
	usedDisk := vals[4]
	totalNet := vals[5]
	usedNet := vals[6]

	if loadAvg > 30 {
		fmt.Printf("Load Average is too high: %d\n", loadAvg)
	}

	if totalMem > 0 {
		memPercent := usedMem * 100 / totalMem
		if memPercent > 80 {
			fmt.Printf("Memory usage too high: %d%%\n", memPercent)
		}
	}

	if totalDisk > 0 {
		diskPercent := usedDisk * 100 / totalDisk
		if diskPercent > 90 {
			freeMb := (totalDisk - usedDisk) / (1024 * 1024)
			fmt.Printf("Free disk space is too low: %d Mb left\n", freeMb)
		}
	}

	if totalNet > 0 {
		netPercent := usedNet * 100 / totalNet
		if netPercent > 90 {
			freeMbit := (totalNet - usedNet) / 1000 / 1000
			fmt.Printf("Network bandwidth usage high: %d Mbit/s available\n", freeMbit)
		}
	}
}
