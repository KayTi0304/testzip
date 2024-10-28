package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"trading/algorithm"
	"trading/backtest"
	"trading/common"

	"github.com/adshao/go-binance/v2"
	"github.com/adshao/go-binance/v2/futures"
	"github.com/aws/aws-lambda-go/lambda"
)

type Request struct {
	UserID    string `json:"userID"`
	ApiKey    string `json:"apiKey"`
	SecretKey string `json:"secretKey"`
}

var (
	client          *futures.Client
	prevDaysFromNow = 31
	investAmt       = 500
	investPercent   = 0.025
	maxLeverage     = 40
	// investPercent = 0.01
	// maxLeverage   = 100

	cutLossMtpier = 1000.0
	// stoplossNoLine   = 3
	filteredMaterial [][]interface{}
)

// func fetchKlinesWithRetry(client *futures.Client, symbol string, interval string) ([]*futures.Kline, error) {
// 	var klines []*futures.Kline
// 	var err error
// 	for retries := 0; retries < 5; retries++ { // Adjust the number of retries as needed
// 		klines, err = common.FetchBinanceFuturesKline(client, symbol, interval)
// 		if err == nil {
// 			return klines, nil
// 		}
// 		log.Printf("Error fetching data for %s: %v. Retrying...", symbol, err)
// 		time.Sleep(time.Duration(retries*2) * time.Second)
// 	}
// 	return nil, fmt.Errorf("failed to fetch klines after retries: %w", err)
// }

func Handler(ctx context.Context, request Request) (common.Response, error) {
	secret, err := common.GetHostApiSecret()
	fmt.Println("1")
	if err != nil {
		return common.Response{
			Code:    400,
			Message: err.Error(),
			Data:    nil,
		}, nil
	}

	client = binance.NewFuturesClient(secret.APIKey, secret.SecretKey)
	//client = binance.NewFuturesClient("LsZXqBiSvmiZ23vEmQpTzQl0khZlxUWCsNmqN3Hp5EbuMCjU1pHjxaLjOQc5VCZg", "YaDdNJ3pn8isRJMkaSpm08rnJhIG0FSj3afMjcHK1aTOa2sORe0kpw6fC9BT55fY")

	whitelist, err := common.GetActiveUSDTFuturesPairs(client)
	if err != nil {
		log.Fatalf("Failed to get active USDT futures pairs: %v", err)
	}
	log.Println(len(whitelist))

	// tpList := []int{15, 20, 25, 30, 40, 50, 55, 60, 70}
	// tpList := []int{21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40}

	// symbol: THETAUSDT|interval: 1h|tp: 45.0|multiplier21: 1.3|multiplier14: 1.8|multiplier10: 2.3|lev: 50|ivp: 0.040|alloc: 500|usdt: 1000.0
	// symbol: VETUSDT|interval: 1h|tp: 65.0|multiplier21: 1.6|multiplier14: 2.6|multiplier10: 3.1|lev: 50|ivp: 0.040|alloc: 500|usdt: 1000.0
	whitelist = []string{"SOLUSDT"}
	// whitelist = []string{"BTCUSDT", "ETHUSDT"}
	// intervals := []string{"1h"}
	// multiplier21 := []float64{0.6}
	// multiplier14 := []float64{1.5}
	// multiplier10 := []float64{2.0}
	tpList := []float64{50.0}

	intervals := []string{"5m"}
	// tpList := common.GenerateSequence(12.0, 41.0, 2.0)
	//tpList := common.GenerateSequence(30.0, 101.0, 5.0)
	//tpList := common.GenerateSequence(90.0, 151.0, 5.0)

	// // tpList := common.GenerateSequence(30.0, 61.0, 5.0)

	// // actLossList := common.GenerateSequence(100.0, 201.0, 10.0)

	// multiplier21 := common.GenerateSequence(0.5, 2.5, 0.1)
	// multiplier14 := common.GenerateSequence(1.0, 3.0, 0.1)
	// multiplier10 := common.GenerateSequence(1.5, 3.5, 0.1)

	//mssPeriod := []int{5, 10, 15, 20, 13}
	//sameTrendPeriod := []int{5, 10, 15, 7, 8}
	//stopLossPeriod := []int{5, 10, 15, 20, 25, 30, 35, 33}
	//minimumTrendCounter := []int{2, 3, 4, 5, 6, 7}
	//displacementCounter := []int{2, 3, 4, 5, 6, 7}

	mssPeriod := []int{14}
	sameTrendPeriod := []int{14}
	stopLossPeriod := []int{35}
	minimumTrendCounter := []int{2}
	displacementCounter := []int{3}
	liquidityPeriod := []int{15}
	cooldownPeriod := []int{5}

	// mssPeriod := []int{18}
	// sameTrendPeriod := []int{14}
	// stopLossPeriod := []int{35}
	// minimumTrendCounter := []int{2}
	// displacementCounter := []int{4}

	// mssPeriod := common.GenerateSequenceInt(10, 25, 2)
	// sameTrendPeriod := common.GenerateSequenceInt(10, 15, 2)
	// stopLossPeriod := common.GenerateSequenceInt(5, 35, 5)
	// minimumTrendCounter := common.GenerateSequenceInt(2, 7, 1)
	// displacementCounter := common.GenerateSequenceInt(2, 7, 1)
	// liquidityPeriod := common.GenerateSequenceInt(15, 40, 5)
	// cooldownPeriod := common.GenerateSequenceInt(5, 20, 1)

	raws := [][]interface{}{
		common.IntToInterfaceSlice(mssPeriod),
		common.IntToInterfaceSlice(sameTrendPeriod),
		common.IntToInterfaceSlice(stopLossPeriod),
		common.IntToInterfaceSlice(minimumTrendCounter),
		common.IntToInterfaceSlice(displacementCounter),
		common.IntToInterfaceSlice(liquidityPeriod),
		common.IntToInterfaceSlice(cooldownPeriod),
		common.FloatToInterfaceSlice(tpList),
		// common.FloatToInterfaceSlice(multiplier21),
		// common.FloatToInterfaceSlice(multiplier14),
		// common.FloatToInterfaceSlice(multiplier10),
	}

	material := common.Product(raws)
	for _, v := range material {
		x := v[0].(int)
		y := v[1].(int)
		a := v[2].(int)
		b := v[3].(int)
		c := v[4].(int)
		d := v[5].(int)
		e := v[6].(int)
		z := v[7].(float64)
		filteredMaterial = append(filteredMaterial, []interface{}{x, y, a, b, c, d, e, z})
		// if z >= y+1 {
		// 	filteredMaterial = append(filteredMaterial, []interface{}{x, y, z})
		// }
	}
	count := len(filteredMaterial) * len(whitelist) * len(intervals)
	log.Println(count)
	resultsBatch := []map[string]interface{}{}
	//batchSize := 10
	//fileIndex := 0

	// maxLeverage := 100
	actualTrade := investPercent * float64(investAmt) * float64(maxLeverage)
	log.Println(actualTrade)
	//var mu sync.Mutex
	//var wg sync.WaitGroup
	//i := 0

	// for _, symbol := range whitelist {
	// Process whitelist in batches of 10
	// batchSize := 10
	// batchCount := 0

	// for i := 0; i < len(whitelist); i += batchSize {
	// 	end := i + batchSize
	// 	if end > len(whitelist) {
	// 		end = len(whitelist)
	// 	}
	// 	batch := whitelist[i:end]

	// 	log.Printf("Processing batch %d with %d items", batchCount+1, len(batch))

	// 	var wg sync.WaitGroup
	// 	var mu sync.Mutex
	// 	res := []map[string]interface{}{}

	for _, symbol := range whitelist {
		log.Println(symbol)
		for _, interval := range intervals {
			klines, err := common.FetchKlinesWithRetry(client, symbol, interval, prevDaysFromNow)
			if err != nil {
				log.Printf("Error fetching data for %s on interval %s: %v", symbol, interval, err)
				continue
			}

			for i, item := range filteredMaterial {
				log.Printf("%d/%d", i, len(filteredMaterial))

				mssP := item[0].(int)
				stP := item[1].(int)
				slP := item[2].(int)
				mtC := item[3].(int)
				dC := item[4].(int)
				lP := item[5].(int)
				cdP := item[6].(int)
				tp := item[7].(float64)

				// mr21 := item[1].(float64)
				// mr14 := item[2].(float64)
				// mr10 := item[3].(float64)

				res := algorithm.TrendsICT(klines, mssP, stP, slP, mtC, dC, lP, cdP)

				title := fmt.Sprintf("symbol: %s|lP: %s|cdP: %s|interval: %s|mssP: %d|stP: %d|slP: %d|mtC: %d|dC: %d|tp: %.1f|lev: %d|ivp: %.3f|alloc: %d|usdt: %.1f", symbol, lP, cdP, interval, mssP, stP, slP, mtC, dC, tp, maxLeverage, investPercent, investAmt, actualTrade)

				_, summaryMetrics := backtest.Backtestv2(res, float64(maxLeverage), tp, investAmt, investPercent, cutLossMtpier)

				fees := float64(summaryMetrics.TotalTrades) * float64(actualTrade) * 0.001
				netPnl := summaryMetrics.TotalPNL - fees
				finalBal := float64(investAmt) + netPnl
				resultsBatch = append(resultsBatch, map[string]interface{}{
					"Symbol":          symbol,
					"Item":            title,
					"Wins":            summaryMetrics.Wins,
					"Losses":          summaryMetrics.Losses,
					"Total Trades":    summaryMetrics.TotalTrades,
					"Win Rate (%)":    summaryMetrics.WinRate,
					"Total Gross PNL": summaryMetrics.TotalPNL,
					"Total Net PNL":   netPnl,
					"Fees":            fees,
					"Final Balance":   finalBal,
				})
				furtherProcess(resultsBatch)

				//common.SaveCsv4(tr.ToMapSlice(), "tr")

				// Save the results for this batch
				//resultsBatch = []map[string]interface{}{} // Reset the batch
				// wg.Add(1)                                                                              // Add to the WaitGroup just before launching the goroutine
				// go func(symbol string, interval string, item []interface{}, klines []*futures.Kline) { // Assuming klines has a type
				// 	defer wg.Done()

				// 	// Safe concurrent access to i
				// 	mu.Lock()
				// 	currentIndex++
				// 	// currentIndex := i + 1
				// 	i++
				// 	mu.Unlock()

				// 	log.Printf("%d/%d", currentIndex, count)
				// 	//tp := item[0].(float64)
				// 	mr21 := item[1].(float64)
				// 	mr14 := item[2].(float64)
				// 	mr10 := item[3].(float64)
				// 	//title := fmt.Sprintf("symbol: %s|interval: %s|tp: %.1f|multiplier21: %.1f|multiplier14: %.1f|multiplier10: %.1f|lev: %d|ivp: %.3f|alloc: %d|usdt: %.1f", symbol, interval, tp, mr21, mr14, mr10, maxLeverage, investPercent, investAmt, actualTrade)

				// 	algorithm.HaThreeSupertrendsV4(klines, mr21, mr14, mr10, 21, 14, 10, stoplossNoLine)
				// 	// resultData := algorithm.HaThreeSupertrendsEx(klines, mr21, mr14, mr10, 21, 14, 10, stoplossNoLine)
				// 	// resultData := algorithm.HaThreeSupertrendsEMA200(klines, mr21, mr14, mr10, 21, 14, 10, stoplossNoLine)
				// 	// resultData := algorithm.HaThreeSupertrendsStochRsi(klines, mr21, mr14, mr10, 21, 14, 10, 14, stoplossNoLine)
				// 	// resultData := algorithm.HaThreeSupertrendsMacd(klines, mr21, mr14, mr10, 21, 14, 10, 12, 26, 9, stoplossNoLine)
				// 	// common.SaveCsv4(resultData.ToMapSlice(), "ALGO")
				// 	/*_, summaryMetrics := algorithm.Backtestv2(resultData, float64(maxLeverage), tp, investAmt, investPercent, cutLossMtpier)
				// 	// common.SaveCsv4(res.ToMapSlice(), "detail")
				// 	// resultData := algorithm.HaTwoSupertrends(klines, mr21, mr14, 21, 14)
				// 	// _, summaryMetrics := algorithm.Backtestv1(resultData, float64(maxLeverage), int(tp), investAmt, investPercent)

				// 	fees := float64(summaryMetrics.TotalTrades) * float64(actualTrade) * 0.001
				// 	netPnl := summaryMetrics.TotalPNL - fees
				// 	finalBal := float64(investAmt) + netPnl*/
				// 	mu.Lock()
				// 	resultsBatch = append(resultsBatch, map[string]interface{}{
				// 		"Symbol":          symbol,
				// 		"Item":            title,
				// 		"Wins":            summaryMetrics.Wins,
				// 		"Losses":          summaryMetrics.Losses,
				// 		"Total Trades":    summaryMetrics.TotalTrades,
				// 		"Win Rate (%)":    summaryMetrics.WinRate,
				// 		"Total Gross PNL": summaryMetrics.TotalPNL,
				// 		"Total Net PNL":   netPnl,
				// 		"Fees":            fees,
				// 		"Final Balance":   finalBal,
				// 	})
				// 	mu.Unlock()
				// }(symbol, interval, item, klines) // Pass klines as an argument to the goroutine
			}

			// Check if we need to save the current batch and reset
			/*if (index+1)%batchSize == 0 || index == len(whitelist)-1 {
				// Wait for all goroutines of this batch to finish
				wg.Wait()
				common.SaveCsv4(resultsBatch, fmt.Sprintf("test-%d-", fileIndex)) // need to change

				furtherProcess(resultsBatch)

				// Save the results for this batch
				// go saveResults(resultsBatch, fmt.Sprintf("results_part_%d.csv", fileIndex))
				resultsBatch = []map[string]interface{}{} // Reset the batch
				fileIndex++
			}*/
		}
	}
	common.SaveCsv4(resultsBatch, fmt.Sprintf("test-%d-", 1))
	// wg.Wait()
	// common.SaveCsv4(res, "test-" + string(fileIndex) + "-")

	return common.Response{
		Code:    200,
		Message: "Request successfully",
		Data:    nil,
	}, nil
}

func furtherProcess(resultsBatch []map[string]interface{}) {
	file, err := os.Open("trading-pairs.json")
	if err != nil {
		log.Fatalf("Failed to open trading-pairs.json: %v", err)
	}
	defer file.Close()

	byteValue, _ := io.ReadAll(file)

	var tradePairs map[string][]map[string]interface{}
	if err := json.Unmarshal(byteValue, &tradePairs); err != nil {
		log.Fatalf("Failed to unmarshal trading-pairs.json: %v", err)
	}

	symbolMaxPnl := make(map[string]map[string]interface{})

	// Iterate through the resultsBatch to find the max PNL item for each symbol
	for _, result := range resultsBatch {
		symbol := result["Symbol"].(string)
		netPnl := result["Total Net PNL"].(float64)
		if existingResult, exists := symbolMaxPnl[symbol]; !exists || netPnl > existingResult["Total Net PNL"].(float64) {
			symbolMaxPnl[symbol] = result
		}
	}
	// for _, result := range resultsBatch {
	// 	symbol := result["Symbol"].(string)
	// 	netPnl := result["Win Rate (%)"].(float64)
	// 	if existingResult, exists := symbolMaxPnl[symbol]; !exists || netPnl > existingResult["Win Rate (%)"].(float64) {
	// 		symbolMaxPnl[symbol] = result
	// 	}
	// }

	// Map the top PNL item to the output function_argument for each symbol
	for symbol, topItem := range symbolMaxPnl {
		// Extract function_argument from the title
		parts := strings.Split(topItem["Item"].(string), "|")
		argsMap := make(map[string]string)
		for _, part := range parts {
			kv := strings.Split(part, ": ")
			if len(kv) == 2 {
				argsMap[kv[0]] = kv[1]
			}
		}
		functionArgument := fmt.Sprintf("{\"mssPeriod\":\"%s\",\"sameTrendPeriod\":\"%s\",\"stopLossPeriod\":\"%s\",\"minimumTrendCounter\":\"%s\",\"displacementCounter\":\"%s\",\"leverage\":\"%s\",\"investVolPercent\":\"%s\",\"takeProfitPercent\":\"%s\"}",
			argsMap["mssP"], argsMap["stP"], argsMap["slP"], argsMap["mtC"], argsMap["dC"], argsMap["lev"], argsMap["ivp"], argsMap["tp"])
		log.Printf("Symbol: %s", symbol)
		log.Printf("WinRate: %s", topItem["Win Rate (%)"])
		log.Printf("NetPnl: %s", topItem["Total Net PNL"])
		log.Printf("Total Trades: %s", topItem["Total Trades"])
		log.Printf("FunctionArgument: %s", functionArgument)

		for i, tradePair := range tradePairs["trade_pairs"] {
			if tradePair["trade_pair"] == symbol {
				var existingArgs map[string]interface{}
				json.Unmarshal([]byte(tradePair["function_argument"].(string)), &existingArgs)
				// existingArgs["multiplier21"] = argsMap["multiplier21"]
				// existingArgs["multiplier14"] = argsMap["multiplier14"]
				// existingArgs["multiplier10"] = argsMap["multiplier10"]
				// existingArgs["leverage"] = argsMap["lev"]
				existingArgs["takeProfitPercent"] = argsMap["tp"]
				newFunctionArgument, _ := json.Marshal(existingArgs)
				tradePairs["trade_pairs"][i]["function_argument"] = string(newFunctionArgument)
			}
		}
	}
	updatedFile, err := json.MarshalIndent(tradePairs, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal updated trade pairs: %v", err)
	}
	if err := os.WriteFile("trading-pairs.json", updatedFile, 0644); err != nil {
		log.Fatalf("Failed to write updated trade pairs to file: %v", err)
	}
}

func main() {
	// lambda.Start(Handler)
	if _, exists := os.LookupEnv("_LAMBDA_SERVER_PORT"); exists {
		lambda.Start(Handler)
	} else {
		// Running locally
		var req Request
		result, err := Handler(context.Background(), req)
		if err != nil {
			log.Fatalf("Error: %v", err)
		} else {
			fmt.Println("Result:", result)
		}
	}
}
