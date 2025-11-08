package timer

import (
	"fmt"
	"fullcycle-auction_go/configuration/logger"
	"os"
	"strconv"
	"time"
)

func AuctionTimer() (time.Duration, error) {

	durationStr := os.Getenv("AUCTION_INTERVAL")
	if durationStr == "" {
		logger.Info("Auction duration from enviroment is empty")
		return time.Duration(time.Now().Day()), fmt.Errorf("auction duration from enviroment is empty")
	}

	seconds, err := strconv.Atoi(durationStr)
	if err != nil {
		logger.Error("Error to parse auction duration to int", err)
		return time.Duration(time.Now().Day()), err
	}

	return time.Duration(seconds) * time.Second, nil
}
