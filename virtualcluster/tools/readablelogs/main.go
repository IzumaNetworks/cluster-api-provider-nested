package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

// LogEntry defines the structure of the JSON log entry
type LogEntry struct {
	Level              string  `json:"level"`
	Timestamp          float64 `json:"ts"`
	Logger             string  `json:"logger"`
	Message            string  `json:"msg"`
	ClusterVersionName string  `json:"ClusterVersionName"`
}

// ANSI color codes
const (
	Reset        = "\033[0m"
	Red          = "\033[31m"
	Green        = "\033[32m"
	Yellow       = "\033[33m"
	Blue         = "\033[34m"
	Magenta      = "\033[35m"
	Cyan         = "\033[36m"
	White        = "\033[37m"
	DarkGrey     = "\033[90m"
	Grey         = "\033[37m"
	UnderlineOn  = "\033[4m"
	UnderlineOff = "\033[24m"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		var entry LogEntry
		var entryMap map[string]interface{}

		err := json.Unmarshal([]byte(line), &entry)
		if err != nil {
			// Print the line as is if it cannot be parsed as JSON
			fmt.Println(line)
			continue
		}

		err = json.Unmarshal([]byte(line), &entryMap)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing JSON to map: %v\n", err)
			continue
		}

		// Determine the color for the log level
		levelColor := Green
		if entry.Level == "error" {
			levelColor = Red
		}

		// Determine if the logger field should be underlined
		logger := entry.Logger
		if strings.HasPrefix(logger, "DEBUG-VC") {
			logger = UnderlineOn + logger + Reset
		}

		// Convert the timestamp to a human-readable date
		ts := time.Unix(int64(entry.Timestamp), int64((entry.Timestamp-float64(int64(entry.Timestamp)))*1e9))
		localTime := ts.Local().Format("2006-01-02 15:04:05 MST")

		// Print the log entry in column-based format with ANSI colors
		fmt.Printf("%s%s%-10s%s | %sTS: %s%-25s%s | %sL: %s%-40s%s | %sM: %s%-30s%s | %sCVN: %s%s%s",
			DarkGrey, levelColor, entry.Level, Reset,
			DarkGrey, White, localTime, Reset,
			DarkGrey, Cyan, logger, Reset,
			DarkGrey, White, entry.Message, Reset,
			DarkGrey, Grey, entry.ClusterVersionName, Reset,
		)

		// Print any unexpected fields
		unexpectedFields := getUnexpectedFields(entryMap, entry)
		if len(unexpectedFields) > 0 {
			fmt.Printf(" | %s--> %s%s", DarkGrey, White, unexpectedFields, Reset)
		}

		fmt.Println()
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading standard input: %v\n", err)
	}
}

// getUnexpectedFields returns a formatted string of unexpected fields in the JSON map
func getUnexpectedFields(entryMap map[string]interface{}, entry LogEntry) string {
	var unexpected []string
	for key, value := range entryMap {
		switch key {
		case "level", "ts", "logger", "msg", "ClusterVersionName":
			// Known fields, do nothing
		default:
			unexpected = append(unexpected, fmt.Sprintf("%s: %v", key, value))
		}
	}
	return strings.Join(unexpected, ", ")
}
