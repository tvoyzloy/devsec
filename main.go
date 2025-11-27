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
		resp, err := http.Get(statsURL)
		if err != nil {
			errorCount = handleError(errorCount, false)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			errorCount = handleError(errorCount, false)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			errorCount = handleError(errorCount, false)
			continue
		}

		values, ok := parseStats(string(body))
		if !ok {
			errorCount = handleError(errorCount, false)
			continue
		}

		// успешный сценарий → сбрасываем счётчик и печатаем показатели
		errorCount = 0
		report(values)
	}
}

func handleError(count int, force bool) int {
	count++
	if count >= 3 || force {
		fmt.Println("Unable to fetch server statistic")
		return 0
	}
	return count
}

func parseStats(raw string) ([7]int64, bool) {
	var values [7]int64

	parts := strings.Split(strings.TrimSpace(raw), ",")
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
			freeMbit := (totalNet - usedNet) * 8 / 1024 / 1024
			fmt.Printf("Network bandwidth usage high: %d Mbit/s available\n", freeMbit)
		}
	}
}
