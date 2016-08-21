package hostchecker

import (
	"strconv"
	"log"
	"regexp"
)

var (
	psTimeRegex = regexp.MustCompile(`((\d+)-)?((\d{2}):)?(\d{2}):(\d{2})`)
)

func extractTimeInSeconds(elapsedTime string) int {
	toIntOrZero := func (s string) int {
		if s == "" {
			return 0
		}
		i, err := strconv.Atoi(s)
		if err != nil {
			log.Fatalf("Failed to convert string to integer: %v", err)
		}
		return i
	}

	if !psTimeRegex.MatchString(elapsedTime) {
		log.Fatalf("Could not match time format '%s'\n", elapsedTime)
	}
	found := psTimeRegex.FindStringSubmatch(elapsedTime)
	days := found[2]
	hours := found[4]
	minutes := found[5]
	seconds := found[6]

	var secondsTotal int
	if days == "" {
		secondsTotal = toIntOrZero(hours) * 60 * 60 + toIntOrZero(minutes) * 60 + toIntOrZero(seconds)
	} else {
		secondsTotal = toIntOrZero(days) * 86400 + toIntOrZero(hours) * 60 * 60 + toIntOrZero(minutes) * 60 + toIntOrZero(seconds)
	}
	//fmt.Printf("%s is %d seconds\n", elapsedTime, secondsTotal)
	return secondsTotal
}


