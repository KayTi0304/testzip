package common

import (
	"encoding/csv"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"time"

	"github.com/adshao/go-binance/v2/futures"
)

type Candle struct {
	OpenTime  string
	Open      float64
	High      float64
	Low       float64
	Close     float64
	CloseTime string
}

type HeikenAshiCandle struct {
	Open, High, Low, Close float64
}

func SaveCsv4(res []map[string]interface{}, fileName string) {
	// Create a new CSV file
	currentTime := time.Now()
	formattedTime := currentTime.Format("2006-01-02-15-04-05")

	// Create the filename with the formatted date and time
	path := fmt.Sprintf("C:/Users/Admin/Documents/trading/%s-%s.csv", fileName, formattedTime) // change path
	log.Println(path)
	file, err := os.Create(path)
	if err != nil {
		log.Fatalln("Failed to create file", err)
	}
	defer file.Close()

	// Create a CSV writer
	writer := csv.NewWriter(file)
	defer writer.Flush()

	if len(res) > 0 {
		// Write the header
		var header []string
		for key := range res[0] {
			header = append(header, key)
		}
		if err := writer.Write(header); err != nil {
			log.Fatalln("Error writing header to CSV", err)
		}

		// Write the records
		for _, record := range res {
			var row []string
			for _, key := range header {
				switch v := record[key].(type) {
				case string:
					row = append(row, v)
				case int:
					row = append(row, strconv.Itoa(v))
				case float64:
					row = append(row, strconv.FormatFloat(v, 'f', -1, 64))
				default:
					row = append(row, "")
				}
			}
			if err := writer.Write(row); err != nil {
				log.Fatalln("Error writing record to CSV", err)
			}
		}
	}

	log.Println("CSV file created successfully at:", path)
}

func GenerateSequenceInt(start, end, step int) []int {
	// Verify that step is positive and non-zero to avoid infinite loops
	if step <= 0 {
		fmt.Println("Step must be a positive number")
		return nil
	}

	// Check if the range is valid
	if start > end {
		fmt.Println("Start value must be less than end value")
		return nil
	}

	// Calculate the number of steps needed (safely handling floating point precision)
	var sequence []int
	for value := start; value <= end; value += step {
		sequence = append(sequence, value)
	}

	return sequence
}

func GenerateSequence(start, end, step float64) []float64 {
	// Verify that step is positive and non-zero to avoid infinite loops
	if step <= 0 {
		fmt.Println("Step must be a positive number")
		return nil
	}

	// Check if the range is valid
	if start > end {
		fmt.Println("Start value must be less than end value")
		return nil
	}

	// Calculate the number of steps needed (safely handling floating point precision)
	var sequence []float64
	for value := start; value <= end; value += step {
		sequence = append(sequence, value)
	}

	return sequence
}

func IntToInterfaceSlice(floats []int) []interface{} {
	result := make([]interface{}, len(floats))
	for i, v := range floats {
		result[i] = v
	}
	return result
}

func FloatToInterfaceSlice(floats []float64) []interface{} {
	result := make([]interface{}, len(floats))
	for i, v := range floats {
		result[i] = v
	}
	return result
}

func Product(raws [][]interface{}) [][]interface{} {
	if len(raws) == 0 {
		return nil
	}
	if len(raws) == 1 {
		result := make([][]interface{}, len(raws[0]))
		for i, v := range raws[0] {
			result[i] = []interface{}{v}
		}
		return result
	}
	prev := Product(raws[:len(raws)-1])
	result := make([][]interface{}, 0)
	for _, x := range prev {
		for _, y := range raws[len(raws)-1] {
			comb := append([]interface{}{}, x...)
			comb = append(comb, y)
			result = append(result, comb)
		}
	}
	return result
}

func CalculateATR(high, low, close []float64, period int) []float64 {
	trueRange := CalculateTrueRange(high, low, close)
	// atr := CalculateEMA(trueRange, period)
	alpha := 1.0 / float64(period)
	atr := calculateEWMA(trueRange, alpha, period)
	return atr
}

func ExtractCloses(candles []HeikenAshiCandle) []float64 {
	closes := make([]float64, len(candles))
	for i, candle := range candles {
		closes[i] = candle.Close
	}
	return closes
}

func ExtractHighs(candles []HeikenAshiCandle) []float64 {
	highs := make([]float64, len(candles))
	for i, candle := range candles {
		highs[i] = candle.High
	}
	return highs
}

func ExtractLows(candles []HeikenAshiCandle) []float64 {
	lows := make([]float64, len(candles))
	for i, candle := range candles {
		lows[i] = candle.Low
	}
	return lows
}

func calculateEWMA(values []float64, alpha float64, minPeriods int) []float64 {
	n := len(values)
	ewma := make([]float64, n)

	// Initialize the first value
	var sum float64
	var count int
	for i := 0; i < n; i++ {
		sum += values[i]
		count++
		if count == minPeriods {
			ewma[i] = sum / float64(minPeriods)
			break
		}
	}

	// Calculate EWMA for the rest of the values
	for i := minPeriods; i < n; i++ {
		ewma[i] = alpha*values[i] + (1-alpha)*ewma[i-1]
	}

	return ewma
}

func CalculateTrueRange(high, low, close []float64) []float64 {
	n := len(high)
	trueRange := make([]float64, n)

	for i := 1; i < n; i++ {
		diff1 := high[i] - low[i]
		diff2 := high[i] - close[i-1]
		diff3 := close[i-1] - low[i]

		maxDiff := math.Max(math.Abs(diff1), math.Abs(diff2))
		maxDiff = math.Max(maxDiff, math.Abs(diff3))

		trueRange[i] = maxDiff
	}

	return trueRange
}

func parseStringToFloat(s string) float64 {
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0.0
	}
	return val
}

func ExtractCandles(klines []*futures.Kline) []Candle {
	candles := make([]Candle, len(klines))
	for i, kline := range klines {
		openTime := time.UnixMilli(kline.OpenTime).UTC().Format("2006-01-02 15:04:05")
		closeTime := time.UnixMilli(kline.CloseTime).UTC().Format("2006-01-02 15:04:05")

		candles[i] = Candle{
			OpenTime:  openTime,
			Open:      parseStringToFloat(kline.Open),
			High:      parseStringToFloat(kline.High),
			Low:       parseStringToFloat(kline.Low),
			Close:     parseStringToFloat(kline.Close),
			CloseTime: closeTime,
		}
	}
	return candles
}

func CalculateHeikenAshi(candles []Candle) []HeikenAshiCandle {
	var haCandles []HeikenAshiCandle

	for i, candle := range candles {
		var haCandle HeikenAshiCandle
		if i == 0 {
			haCandle.Open = (candle.Open + candle.Close) / 2
		} else {
			haCandle.Open = (haCandles[i-1].Open + haCandles[i-1].Close) / 2
		}
		haCandle.Close = (candle.Open + candle.High + candle.Low + candle.Close) / 4
		haCandle.High = max(candle.High, haCandle.Open, haCandle.Close)
		haCandle.Low = min(candle.Low, haCandle.Open, haCandle.Close)

		haCandles = append(haCandles, haCandle)
	}

	return haCandles
}
