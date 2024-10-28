package algorithm

import (
	"math"
	"trading/common"

	"github.com/adshao/go-binance/v2/futures"
)

type SupertrendData2 struct {
	OpenTime         []string
	CloseTime        []string
	Open             []float64
	High             []float64
	Low              []float64
	Close            []float64
	HaOpen           []float64
	HaHigh           []float64
	HaLow            []float64
	HaClose          []float64
	Supertrend21     []bool
	FinalLowerBand21 []float64
	FinalUpperBand21 []float64
	Supertrend14     []bool
	FinalLowerBand14 []float64
	FinalUpperBand14 []float64
	Supertrend10     []bool
	FinalLowerBand10 []float64
	FinalUpperBand10 []float64
	Rsi21            []float64
	Rsi3             []float64
	AdxSignal        []float64
	BUY              []float64
	SELL             []float64
	StopLoss         []float64
}

func (data *SupertrendData2) ToMapSlice() []map[string]interface{} {
	length := len(data.Supertrend21) // Assuming all slices are the same length
	results := make([]map[string]interface{}, length)

	for i := 0; i < length; i++ {
		row := map[string]interface{}{
			"OpenTime":         data.OpenTime[i],
			"CloseTime":        data.CloseTime[i],
			"Open":             data.Open[i],
			"High":             data.High[i],
			"Low":              data.Low[i],
			"Close":            data.Close[i],
			"HaOpen":           data.HaOpen[i],
			"HaHigh":           data.HaHigh[i],
			"HaLow":            data.HaLow[i],
			"HaClose":          data.HaClose[i],
			"Supertrend21":     data.Supertrend21[i],
			"FinalLowerBand21": data.FinalLowerBand21[i],
			"FinalUpperBand21": data.FinalUpperBand21[i],
			"Supertrend14":     data.Supertrend14[i],
			"FinalLowerBand14": data.FinalLowerBand14[i],
			"FinalUpperBand14": data.FinalUpperBand14[i],
			"Supertrend10":     data.Supertrend10[i],
			"FinalLowerBand10": data.FinalLowerBand10[i],
			"FinalUpperBand10": data.FinalUpperBand10[i],
			"BUY":              data.BUY[i],
			"SELL":             data.SELL[i],
			"StopLoss":         data.StopLoss[i],
		}
		results[i] = row
	}

	return results
}

func isBuyTrend2(supertrend21, supertrend14, supertrend10 []bool, curr, prev int) bool {
	return (supertrend21[curr] && !supertrend21[prev] && supertrend14[curr] && supertrend10[curr]) ||
		(supertrend14[curr] && !supertrend14[prev] && supertrend21[curr] && supertrend10[curr]) ||
		(supertrend10[curr] && !supertrend10[prev] && supertrend21[curr] && supertrend14[curr])
}

func isSellTrend2(supertrend21, supertrend14, supertrend10 []bool, curr, prev int) bool {
	return (!supertrend21[curr] && supertrend21[prev] && !supertrend14[curr] && !supertrend10[curr]) ||
		(!supertrend14[curr] && supertrend14[prev] && !supertrend21[curr] && !supertrend10[curr]) ||
		(!supertrend10[curr] && supertrend10[prev] && !supertrend21[curr] && !supertrend14[curr])
}

func HaThreeSupertrends(klines []*futures.Kline, multiplier21, multiplier14, multiplier10 float64, atrPeriod21, atrPeriod14, atrPeriod10, stoplossNoLine int) SupertrendData2 {
	candles := common.ExtractCandles(klines)
	haCandles := common.CalculateHeikenAshi(candles)
	length := len(haCandles)
	data := SupertrendData2{
		OpenTime:         make([]string, length),
		CloseTime:        make([]string, length),
		Open:             make([]float64, length),
		High:             make([]float64, length),
		Low:              make([]float64, length),
		Close:            make([]float64, length),
		HaOpen:           make([]float64, length),
		HaHigh:           make([]float64, length),
		HaLow:            make([]float64, length),
		HaClose:          make([]float64, length),
		Supertrend21:     make([]bool, length),
		FinalLowerBand21: make([]float64, length),
		FinalUpperBand21: make([]float64, length),
		Supertrend14:     make([]bool, length),
		FinalLowerBand14: make([]float64, length),
		FinalUpperBand14: make([]float64, length),
		Supertrend10:     make([]bool, length),
		FinalLowerBand10: make([]float64, length),
		FinalUpperBand10: make([]float64, length),
		BUY:              make([]float64, length),
		SELL:             make([]float64, length),
		StopLoss:         make([]float64, length),
	}

	// Calculate ATR for both periods
	atr21 := common.CalculateATR(common.ExtractHighs(haCandles), common.ExtractLows(haCandles), common.ExtractCloses(haCandles), atrPeriod21)
	atr14 := common.CalculateATR(common.ExtractHighs(haCandles), common.ExtractLows(haCandles), common.ExtractCloses(haCandles), atrPeriod14)
	atr10 := common.CalculateATR(common.ExtractHighs(haCandles), common.ExtractLows(haCandles), common.ExtractCloses(haCandles), atrPeriod10)

	for i := 1; i < length; i++ {
		// Set OpenTime and CloseTime
		data.OpenTime[i] = candles[i].OpenTime
		data.CloseTime[i] = candles[i].CloseTime

		data.Open[i] = candles[i].Open
		data.High[i] = candles[i].High
		data.Low[i] = candles[i].Low
		data.Close[i] = candles[i].Close

		data.HaOpen[i] = haCandles[i].Open
		data.HaHigh[i] = haCandles[i].High
		data.HaLow[i] = haCandles[i].Low
		data.HaClose[i] = haCandles[i].Close

		hl2 := (haCandles[i].High + haCandles[i].Low) / 2
		data.FinalUpperBand21[i] = hl2 + multiplier21*atr21[i]
		data.FinalLowerBand21[i] = hl2 - multiplier21*atr21[i]

		data.FinalUpperBand14[i] = hl2 + multiplier14*atr14[i]
		data.FinalLowerBand14[i] = hl2 - multiplier14*atr14[i]

		data.FinalUpperBand10[i] = hl2 + multiplier10*atr10[i]
		data.FinalLowerBand10[i] = hl2 - multiplier10*atr10[i]

		// Calculate Supertrend logic for 21 period
		if haCandles[i].Close > data.FinalUpperBand21[i-1] {
			data.Supertrend21[i] = true
		} else if haCandles[i].Close < data.FinalLowerBand21[i-1] {
			data.Supertrend21[i] = false
		} else {
			data.Supertrend21[i] = data.Supertrend21[i-1]
			if data.Supertrend21[i] {
				if data.FinalLowerBand21[i] < data.FinalLowerBand21[i-1] {
					data.FinalLowerBand21[i] = data.FinalLowerBand21[i-1]
				}
			} else {
				if data.FinalUpperBand21[i] > data.FinalUpperBand21[i-1] {
					data.FinalUpperBand21[i] = data.FinalUpperBand21[i-1]
				}
			}
		}
		if data.Supertrend21[i] {
			data.FinalUpperBand21[i] = math.NaN()
		} else {
			data.FinalLowerBand21[i] = math.NaN()
		}

		if haCandles[i].Close > data.FinalUpperBand14[i-1] {
			data.Supertrend14[i] = true
		} else if haCandles[i].Close < data.FinalLowerBand14[i-1] {
			data.Supertrend14[i] = false
		} else {
			data.Supertrend14[i] = data.Supertrend14[i-1]
			if data.Supertrend14[i] {
				if data.FinalLowerBand14[i] < data.FinalLowerBand14[i-1] {
					data.FinalLowerBand14[i] = data.FinalLowerBand14[i-1]
				}
			} else {
				if data.FinalUpperBand14[i] > data.FinalUpperBand14[i-1] {
					data.FinalUpperBand14[i] = data.FinalUpperBand14[i-1]
				}
			}
		}
		if data.Supertrend14[i] {
			data.FinalUpperBand14[i] = math.NaN()
		} else {
			data.FinalLowerBand14[i] = math.NaN()
		}

		if haCandles[i].Close > data.FinalUpperBand10[i-1] {
			data.Supertrend10[i] = true
		} else if haCandles[i].Close < data.FinalLowerBand10[i-1] {
			data.Supertrend10[i] = false
		} else {
			data.Supertrend10[i] = data.Supertrend10[i-1]
			if data.Supertrend10[i] {
				if data.FinalLowerBand10[i] < data.FinalLowerBand10[i-1] {
					data.FinalLowerBand10[i] = data.FinalLowerBand10[i-1]
				}
			} else {
				if data.FinalUpperBand10[i] > data.FinalUpperBand10[i-1] {
					data.FinalUpperBand10[i] = data.FinalUpperBand10[i-1]
				}
			}
		}
		if data.Supertrend10[i] {
			data.FinalUpperBand10[i] = math.NaN()
		} else {
			data.FinalLowerBand10[i] = math.NaN()
		}

		// STOP LOSS
		if !math.IsNaN(data.FinalLowerBand14[i]) && stoplossNoLine == 2 {
			data.StopLoss[i] = data.FinalLowerBand14[i]
		}

		if !math.IsNaN(data.FinalUpperBand14[i]) && stoplossNoLine == 2 {
			data.StopLoss[i] = data.FinalUpperBand14[i]
		}

		if !math.IsNaN(data.FinalLowerBand10[i]) && stoplossNoLine == 3 {
			data.StopLoss[i] = data.FinalLowerBand10[i]
		}

		if !math.IsNaN(data.FinalUpperBand10[i]) && stoplossNoLine == 3 {
			data.StopLoss[i] = data.FinalUpperBand10[i]
		}

		// Check for buy or sell signals
		if isBuyTrend2(data.Supertrend21, data.Supertrend14, data.Supertrend10, i, i-1) {
			data.BUY[i] = candles[i].Close
		}
		if isSellTrend2(data.Supertrend21, data.Supertrend14, data.Supertrend10, i, i-1) {
			data.SELL[i] = candles[i].Close
		}
	}

	return data
}
