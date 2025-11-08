package timer

import (
	"os"
	"strconv"
	"time"
)

func AuctionTimer() (time.Duration, error) {

	durationStr := os.Getenv("")
	if durationStr == "" {
		//
		//
	}

	minutes, err := strconv.Atoi(durationStr)
	if err != nil {
		//
		//
	}

	return time.Duration(minutes) * time.Minute, nil
}
