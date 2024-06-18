package main

import (
	"flag"
	"fmt"
	"monitoring-service/db"
	pchttp "monitoring-service/http"
	"monitoring-service/provider/pocket"
	"net/http"
	"os"

	"monitoring-service/monitoring"

	"git.mills.io/prologic/bitcask"
	"github.com/go-kit/kit/log"
)

const (
	batchSize        = 8
	defaultPocketURL = "https://mainnet.gateway.pokt.network/v1/lb/61d4a60d431851003b628aa8/v1"
)

func main() {
	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	defaultDBPath, err := os.Getwd()
	if err != nil {
		_ = logger.Log("ERROR: failed to get working directory")
		panic(err)
	}

	dbPath := flag.String("dbPath", defaultDBPath+"/../.pokt-calculator-db", "Path to DB data")
	pocketRpcURL := flag.String("pocketURL", defaultPocketURL, "Pocket network RPC URL")
	flag.Parse()

	fmt.Println("Start...")
	clientWithoutLogger := http.Client{}
	httpClient := pchttp.NewClientWithLogger(clientWithoutLogger, logger)

	// db
	_ = logger.Log("bitcask DB", *dbPath)
	bitcaskDB, err := bitcask.Open(*dbPath)
	if err != nil {
		_ = logger.Log("ERROR opening database")
		panic(err)
	}
	defer func(bitcaskDB *bitcask.Bitcask) {
		err := bitcaskDB.Close()
		if err != nil {
			_ = logger.Log("ERROR closing database")
		}
	}(bitcaskDB)
	blockTimesRepo := db.NewBlockTimesRepo(bitcaskDB)
	paramsRepo := db.NewParamsRepo(bitcaskDB)
	nodeRepo := db.NewNodesRepo(bitcaskDB)

	// provider + MonitoringService
	prv := pocket.NewPocketProvider(httpClient, *pocketRpcURL, blockTimesRepo, paramsRepo, nodeRepo)

	pocketProvider := prv.WithLogger(logger)
	nodeSvc := monitoring.NewService(pocketProvider)

	height, err := nodeSvc.Height()
	maxHeight := int(height)

	for i := 1; i <= maxHeight; i++ {
		thisBatchSize := batchSize
		numLeft := maxHeight - i
		if numLeft < batchSize {
			thisBatchSize = numLeft
		}

		fmt.Printf("This batch size: %d\n", thisBatchSize)

		heights := make([]uint, thisBatchSize)
		for j := 0; j < thisBatchSize; j++ {
			heights[j] = uint(i + j)
		}

		(func() {
			fmt.Printf("fetching blocks %v\n", heights)
			if _, err := nodeSvc.BlockTimes(heights); err != nil {
				fmt.Printf("ERROR: %s", err)
			}
			for _, n := range heights {
				if _, err := nodeSvc.ParamsAtHeight(int64(n), false); err != nil {
					fmt.Printf("ERROR: %s", err)
				}
			}
		})()
		i = i + batchSize - 1
	}
}
