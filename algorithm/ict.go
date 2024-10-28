package algorithm

import (
	"math"
	"trading/common"

	"github.com/adshao/go-binance/v2/futures"
)

type ICTData struct {
	OpenTime   []string
	CloseTime  []string
	Open       []float64
	High       []float64
	Low        []float64
	Close      []float64
	SwingHighs []float64
	SwingLows  []float64
	MssBullish []int
	MssBearish []int
	BullishLP  []int
	BearishLP  []int
	BullishOTE []int
	BearishOTE []int
	BUY        []float64
	SELL       []float64
	StopLoss   []float64
}

// ================ LIQUIDITY POOL =============
// CURRENTLY IN USE
func GetLiquidityPool(candles []common.Candle, interval int, waitInterval int) ([]int, []int) {
	n := len(candles)
	bullishLiquidityPool := make([]int, n)
	bearishLiquidityPool := make([]int, n)

	// 1. get swing highs and swing lows
	swingHighs, swingLows := GetSwingHighLows(candles)
	// 2. find old highs and old lows
	oldHighs, oldLows := GetOldHighsOldLows(swingHighs, swingLows, interval)

	i := 0
	for i < n {
		// bullish
		// 3. to find signal for entry point, find a candle that enters the swing low zone
		// 4. find discount zone and entry (find swing low - swing high and mark zones)
		if oldLows[i] > 0 {
			j := i + waitInterval
			for j < i+waitInterval+interval && j < n {
				idx := findDiscountPremiumZone(candles, i, j, true)
				if idx > 0 {
					fillRange(bullishLiquidityPool, i, idx)
					bullishLiquidityPool[idx] = 2
					i = idx
				}
				j++
			}
		}

		// bearish
		// 3. to find signal for entry point, find a candle that enters the swing high zone
		// 4. find discount zone and entry (find swing high - swing low and mark zones)
		if oldHighs[i] > 0 {
			j := i + waitInterval
			for j < i+waitInterval+interval && j < n {
				idx := findDiscountPremiumZone(candles, i, j, false)
				if idx > 0 {
					fillRange(bearishLiquidityPool, i, idx)
					bearishLiquidityPool[idx] = 2
					i = idx
				}
				j++
			}
		}

		i++
	}

	return bullishLiquidityPool, bearishLiquidityPool
}

// finds the entry of the discount/premium zone.
func findDiscountPremiumZone(candles []common.Candle, start, end int, findDiscount bool) int {
	if start >= end || end >= len(candles) {
		return -1
	}

	if findDiscount {
		target := candles[start].Low
		for i := start + 1; i <= end; i++ {
			if candles[i].Low < target {
				return i
			}
		}

		return -1
	} else {
		target := candles[start].High
		for i := start + 1; i <= end; i++ {
			if candles[i].High > target {
				return i
			}
		}

		return -1
	}

}

// finds the old highs (highests of all swing highs) and old lows (lowests of all swing lows) for each interval within the entire slice.
func GetOldHighsOldLows(swingHighs, swingLows []float64, interval int) ([]float64, []float64) {
	n := len(swingHighs)
	oldHighs := make([]float64, n)
	oldLows := make([]float64, n)

	for i := 0; i < n; i += interval {
		// determine old highs
		maxPrice := math.Inf(-1)
		maxPriceIdx := -1

		for j := i; j < i+interval && j < n; j++ {
			if swingHighs[j] > 0 && swingHighs[j] > maxPrice {
				maxPrice = swingHighs[j]
				maxPriceIdx = j
			}
		}

		if maxPriceIdx >= 0 {
			oldHighs[maxPriceIdx] = maxPrice
		}

		lowPrice := math.Inf(1)
		lowPriceIdx := -1

		// determine old low
		for j := i; j < i+interval && j < n; j++ {
			if swingLows[j] > 0 && swingLows[j] < lowPrice {
				lowPrice = swingLows[j]
				lowPriceIdx = j
			}
		}

		// Store the max price for this interval
		if lowPriceIdx >= 0 {
			oldLows[lowPriceIdx] = lowPrice
		}
	}

	return oldHighs, oldLows
}

/*
Calculate MSSs from candles using High, Low, Open Close. This function WILL depend on GetFairValueGaps to determine displacement.

CONDITION:
Bullish MSS: displacement's movement breaks the previous swing high and is bullish displacement

Bearish MSS: displacement's movement breaks the previous swing low and is bearish displacement

Source: https://innercircletrader.net/tutorials/ict-market-structure-shift/

	https://fxopen.com/blog/en/market-structure-shift-meaning-and-use-in-ict-trading/#:~:text=A%20Market%20Structure%20Shift%20(MSS,through%20a%20key%20market%20level.

The MSS is an area; the return array should indicate the area of the MSS. For example, if the start of the MSS is at index 1 and ends at index 3,
the array should look like [false, true, true, true, ...].
  - start of MSS: the low point of the swing high/low
  - end of MSS: when the displacement first breaks the low/high of swing high/low
*/
func GetMarketStructureShift(candles []common.Candle, constant, displacementCounter int) ([]int, []int, []float64, []float64, []float64) { // bearish, bullish
	//TODO:
	// 1. Find swing high (3 candles; middle candle has highest high)
	// 2. Find swing low (3 candles: middle candle has lowest high)
	swingHighs, swingLows := GetSwingHighLows(candles)

	// 3. Get FVG (tbh not necessary also; just need obvious downward/upward trend)
	//fvgBearish, fvgBullish := GetFairValueGaps(candles)

	// 4. Go through each candle, find the MSS area (update the previous highest/lowest point with swing high/low if needed)
	n := len(candles)
	mssBullish := make([]int, n)
	mssBearish := make([]int, n)
	stopLoss := make([]float64, n)

	i := 0
	for i < n {
		if swingHighs[i] > 0 {
			j := i
			for j < i+constant {
				ok, start, breakHigh := isBullishDisplacement(candles, i+1, j, displacementCounter)
				ll := achievedLLInBetween(i+1, start, swingLows)
				if ok && ll > 0 {
					// displacement
					fillRange(mssBullish, i, breakHigh)
					mssBullish[breakHigh] = 2
					stopLoss[breakHigh] = ll
					j = breakHigh
					i = j
					break
				}

				ok, start, breakHigh = IsBullishWithFVG(candles, i+1, j, displacementCounter)
				ll = achievedLLInBetween(i+1, start, swingLows)
				if ok && ll > 0 {
					// using fvg as displacement
					fillRange(mssBullish, i, breakHigh)
					mssBullish[breakHigh] = 2
					stopLoss[breakHigh] = ll
					j = breakHigh
					i = j
					break
				}

				j++
			}
		} else if swingLows[i] > 0 {
			j := i
			for j < i+constant && j < n {

				ok, start, breakLow := isBearishDisplacement(candles, i+1, j, displacementCounter)
				hh := achievedHHInBetween(i+1, start, swingHighs)
				if ok && hh > 0 {
					// displacements
					fillRange(mssBearish, i, breakLow)
					mssBearish[breakLow] = 2
					stopLoss[breakLow] = hh
					j = breakLow
					i = j
					break
				}

				ok, start, breakLow = IsBearishWithFVG(candles, i+1, j, displacementCounter)
				hh = achievedHHInBetween(i+1, start, swingHighs)
				if ok && hh > 0 {
					// using  fvg as displacement
					fillRange(mssBearish, i, breakLow)
					mssBearish[breakLow] = 2
					stopLoss[breakLow] = hh
					j = breakLow
					i = j
					break
				}

				j++
			}
		}
		i++
	}

	return mssBearish, mssBullish, swingHighs, swingLows, stopLoss
}

// checks if found any higher highs between the range.
func achievedHHInBetween(start, end int, swingHighs []float64) float64 {
	foundHH := -9999.99

	for i := start; i <= end; i++ {
		if swingHighs[i] > 0 {
			foundHH = math.Max(foundHH, swingHighs[i])
			break
		}
	}

	if foundHH == -9999.99 {
		return -1
	}
	return foundHH
}

// checks if found any lower lows between the range.
func achievedLLInBetween(start, end int, swingLows []float64) float64 {
	foundLL := 9999.99

	for i := start; i <= end; i++ {
		if swingLows[i] > 0 {
			foundLL = math.Min(foundLL, swingLows[i])
			break
		}
	}

	if foundLL == 9999.99 {
		return -1
	}
	return foundLL
}

// checks if a minimum number of (bearish) displacement candles exists within the range.
// returns the start and index of candle that breaks the low previous found (the point that the market shifted).
func isBearishDisplacement(candles []common.Candle, start, end, displacementCounter int) (bool, int, int) {
	if start >= end || end >= len(candles) {
		return false, -1, -1
	}

	// check if exist displacement between this range
	displacements := []int{}
	prevC := candles[start].Close
	for i := start; i <= end; i++ {
		if !isBullish(candles[i]) && candles[i].Close < prevC {
			displacements = append(displacements, i)
			prevC = candles[i].Close
		}
	}

	if len(displacements) < displacementCounter {
		return false, -1, -1
	}

	// find the break low
	endd := -1
	for _, idx := range displacements {
		if candles[idx].Close > candles[start-1].Low {
			endd = idx
		}
	}

	return endd > 0 && len(displacements) >= displacementCounter, displacements[0], endd
}

// checks if a minimum number of (bullish) displacement candles exists within the range.
// returns the start and the index of candle that breaks the high previous found (the point that the market shifted).
func isBullishDisplacement(candles []common.Candle, start, end, displacementCounter int) (bool, int, int) {
	if start >= end || end >= len(candles) {
		return false, -1, -1
	}

	// check if exist displacement between this range
	displacements := []int{}
	prevC := candles[start].Close
	for i := start; i <= end; i++ {
		if isBullish(candles[i]) && candles[i].Close > prevC {
			displacements = append(displacements, i)
			prevC = candles[i].Close
		}
	}

	if len(displacements) < displacementCounter {
		return false, -1, -1
	}

	// find the break high
	endd := -1
	for _, idx := range displacements {
		if candles[idx].Close > candles[start-1].High {
			endd = idx
		}
	}

	return endd > 0 && len(displacements) >= displacementCounter, displacements[0], endd
}

// checks if a minimum number of (bearish) fvg exists within the range.
// returns the start and index of candle that breaks the low previous found (the point that the market shifted).
func IsBearishWithFVG(candles []common.Candle, start, end, fvgCounter int) (bool, int, int) {
	if start >= end || end >= len(candles) || end-start <= 2 {
		return false, -1, -1
	}

	// check if exist fvg between this range
	fvg := []int{}
	for i := start + 1; i <= end-1; i++ {
		if isFVG(candles, i-1, i, i+1, false) {
			fvg = append(fvg, i)
		}
	}

	if len(fvg) < fvgCounter {
		return false, -1, -1
	}

	// find the break low
	endd := -1
	for _, idx := range fvg {
		if candles[idx].Close > candles[start-1].Low {
			endd = idx
		}
	}

	return endd > 0 && len(fvg) >= fvgCounter, fvg[0], endd
}

// checks if a minimum number of (bullish) fvg exists within the range.
// returns the start and index of candle that breaks the high previous found (the point that the market shifted).
func IsBullishWithFVG(candles []common.Candle, start, end, fvgCounter int) (bool, int, int) {
	if start >= end || end >= len(candles) || end-start <= 2 {
		return false, -1, -1
	}

	// check if exist fvg between this range
	fvg := []int{}
	for i := start + 1; i <= end-1; i++ {
		if isFVG(candles, i-1, i, i+1, true) {
			fvg = append(fvg, i)
		}
	}

	if len(fvg) < fvgCounter {
		return false, -1, -1
	}

	// find the break high
	endd := -1
	for _, idx := range fvg {
		if candles[idx].Close > candles[start-1].High {
			endd = idx
		}
	}

	return endd > 0 && len(fvg) >= fvgCounter, fvg[0], endd
}

// check if its an fvg using 3 candles.
func isFVG(candles []common.Candle, i, j, k int, findBullish bool) bool {
	if findBullish {
		return candles[i].High <= candles[j].High && candles[j].High <= candles[k].High && // ascending
			candles[j].Open < candles[i].High && candles[j].Close > candles[k].Low && // within body
			candles[i].High < candles[k].Low // no overlap
	} else {
		return candles[i].High >= candles[j].High && candles[j].High >= candles[k].High && // descending
			candles[j].Open >= candles[i].Low && candles[j].Close <= candles[k].High && // within body
			candles[i].Low > candles[k].High // no overlap
	}
}

// find all swing highs and swing lows of the candle slice.
func GetSwingHighLows(candles []common.Candle) ([]float64, []float64) {
	n := len(candles)
	swingHighs := make([]float64, n)
	swingLows := make([]float64, n)

	if n < 3 {
		return swingHighs, swingLows
	}

	for i := 1; i < n-1; i++ {
		if //isBullish(candles[i-1]) && isBullish(candles[i]) && !isBullish(candles[i+1]) &&
		candles[i].High >= candles[i-1].High && candles[i].High >= candles[i+1].High {
			// swing high
			swingHighs[i] = candles[i].High
		}

		if //!isBullish(candles[i-1]) && !isBullish(candles[i]) && isBullish(candles[i+1]) &&
		candles[i].Low <= candles[i-1].Low && candles[i].Low <= candles[i+1].Low {
			// swing low
			swingLows[i] = candles[i].Low
		}
	}

	return swingHighs, swingLows
}

// chgeck if a candle is bullish.
func isBullish(candle common.Candle) bool {
	return candle.Close > candle.Open
}

// fill the range with 1s.
func fillRange(slice []int, start, end int) {
	for i := start; i <= end; i++ {
		slice[i] = 1
	}
}

func TrendsICT(klines []*futures.Kline, mssPeriod, sameTrendPeriod, stopLossPeriod, minimumTrendCounter, displacementCounter, liquidityPeriod, cooldownPeriod int) ICTData {
	candles := common.ExtractCandles(klines)
	length := len(candles)
	data := ICTData{
		OpenTime:   make([]string, length),
		CloseTime:  make([]string, length),
		Open:       make([]float64, length),
		High:       make([]float64, length),
		Low:        make([]float64, length),
		Close:      make([]float64, length),
		MssBullish: make([]int, length),
		MssBearish: make([]int, length),
		BullishLP:  make([]int, length),
		BearishLP:  make([]int, length),
		BearishOTE: make([]int, length),
		BullishOTE: make([]int, length),
		SwingHighs: make([]float64, length),
		SwingLows:  make([]float64, length),
		StopLoss:   make([]float64, length),
		BUY:        make([]float64, length),
		SELL:       make([]float64, length),
	}

	mssBearish, mssBullish, swingHighs, swingLows, _ := GetMarketStructureShift(candles, mssPeriod, displacementCounter)
	//stopLossForSell, stopLossForBuy := GetLiquidity(candles, 30)
	//bullishUnicorn, bearishUnicorn, _ := UnicornModel(candles, bbConstant, uniConstant)
	bullishLP, bearishLP := GetLiquidityPool(candles, liquidityPeriod, cooldownPeriod)
	bullishOTE, bearishOTE := GetOptimalTradeEntry(candles, 25, 15, 6)

	currentAction := ""
	//currentSL := 0.0

	for i, candle := range candles {
		data.Open[i] = candle.Open
		data.Close[i] = candle.Close
		data.OpenTime[i] = candle.OpenTime
		data.CloseTime[i] = candle.CloseTime
		data.High[i] = candle.High
		data.Low[i] = candle.Low
		data.MssBearish[i] = mssBearish[i]
		data.MssBullish[i] = mssBullish[i]
		data.SwingHighs[i] = swingHighs[i]
		data.SwingLows[i] = swingLows[i]
		data.BullishLP[i] = bullishLP[i]
		data.BearishLP[i] = bearishLP[i]
		data.BullishOTE[i] = bullishOTE[i]
		data.BearishOTE[i] = bearishOTE[i]

		if currentAction == "SELL" {
			data.StopLoss[i] = GetStopLoss(candles, i, stopLossPeriod, false)
			//data.StopLoss[i] = currentSL
		} else if currentAction == "BUY" {
			data.StopLoss[i] = GetStopLoss(candles, i, stopLossPeriod, true)
			//data.StopLoss[i] = currentSL
		}

		if mssBearish[i] == 2 {
			// find the start of this bearish mss and check if the trend before is upwards
			j := i - 1
			for j >= 0 && mssBearish[j] == 1 {
				j--
			}
			if isSameTrend(j, candles, sameTrendPeriod, minimumTrendCounter, true) {
				data.SELL[i] = candle.Close
				data.StopLoss[i] = GetStopLoss(candles, i, stopLossPeriod, false)
				data.StopLoss[i+1] = GetStopLoss(candles, i+1, stopLossPeriod, false)
				// data.StopLoss[i] = stopLoss[i]
				// data.StopLoss[i+1] = stopLoss[i]
				// currentSL = stopLoss[i]
				currentAction = "SELL"
			}
		}

		if mssBullish[i] == 2 {
			// find the start of this bearish mss and check if the trend before is downwards
			j := i - 1
			for j >= 0 && mssBullish[j] == 1 {
				j--
			}
			if isSameTrend(j, candles, sameTrendPeriod, minimumTrendCounter, false) {
				data.BUY[i] = candle.Close
				data.StopLoss[i] = GetStopLoss(candles, i, stopLossPeriod, true)
				// data.StopLoss[i] = stopLoss[i]
				// currentSL = stopLoss[i]
				currentAction = "BUY"
			}
		}

		if bearishLP[i] == 2 && mssBearish[i] != 2 && mssBullish[i] != 2 {
			data.SELL[i] = candle.Close
			data.StopLoss[i] = GetStopLoss(candles, i, stopLossPeriod, false)
			//data.StopLoss[i+1] = GetStopLoss(candles, i+1, stopLossPeriod, false)
			currentAction = "SELL"
		}

		if bullishLP[i] == 2 && mssBullish[i] != 2 && mssBearish[i] != 2 {
			data.BUY[i] = candle.Close
			data.StopLoss[i] = GetStopLoss(candles, i, stopLossPeriod, true)
			currentAction = "BUY"
		}

		// if bullishOTE[i] == 2 && mssBearish[i] != 2 && mssBullish[i] != 2 && bearishLP[i] != 2 && bullishLP[i] != 2 {
		// 	data.BUY[i] = candle.Close
		// 	data.StopLoss[i] = GetStopLoss(candles, i, stopLossPeriod, true)
		// 	currentAction = "BUY"
		// }

		// if bearishOTE[i] == 2 && mssBearish[i] != 2 && mssBullish[i] != 2 && bearishLP[i] != 2 && bullishLP[i] != 2 {
		// 	data.SELL[i] = candle.Close
		// 	data.StopLoss[i] = GetStopLoss(candles, i, stopLossPeriod, false)
		// 	currentAction = "SELL"
		// }

		// if bearishUnicorn[i] > 0 {
		// 	data.SELL[i] = candle.Close
		// 	data.StopLoss[i] = candles[i].High + 0.3
		// }

		// if bullishUnicorn[i] > 0 {
		// 	data.BUY[i] = candle.Close
		// 	data.StopLoss[i] = candles[i].Low - 0.3
		// }
	}
	common.SaveCsv4(data.ToMapSlice(), "pain")
	return data
}

// ============= HELPFUL FUNCTIONS =================

// get the stop loss for the candle by finding the highest high/lowest low within the range.
func GetStopLoss(candles []common.Candle, idx, period int, findBuy bool) float64 {
	// see forward
	res := 0.0
	if !findBuy {
		res = candles[idx].High
	} else {
		res = candles[idx].Low
	}

	for i := idx; i >= 0 && i >= idx-period; i-- {
		if !findBuy {
			res = math.Max(res, candles[i].High)
		} else {
			res = math.Min(res, candles[i].Low)
		}
	}

	// see backward
	for i := idx; i < len(candles) && i <= idx-period; i++ {
		if !findBuy {
			res = math.Max(res, candles[i].High)
		} else {
			res = math.Min(res, candles[i].Low)
		}
	}

	return res
}

// checks if the range has at least a minimum number of upward/downward candles to form the desired trend.
func isSameTrend(idx int, candles []common.Candle, period, minimumTrendCounter int, detectUptrend bool) bool {
	if idx-period < 0 {
		return false
	}

	trendCounter := 0
	prevC := candles[idx].Close
	if detectUptrend {
		for i := idx - 1; i > idx-period; i-- {
			if prevC >= candles[i].Close { // doing in reverse
				trendCounter++
				prevC = candles[i].Close

			}
		}
	} else {
		for i := idx - 1; i > idx-period; i-- {
			if prevC <= candles[i].Close { // doing in reverse
				trendCounter++
				prevC = candles[i].Close
			}

		}
	}

	return trendCounter >= minimumTrendCounter
}

func (data *ICTData) ToMapSlice() []map[string]interface{} {
	length := len(data.Open) // Assuming all slices are the same length
	results := make([]map[string]interface{}, length)

	for i := 0; i < length; i++ {
		row := map[string]interface{}{
			"OpenTime":  data.OpenTime[i],
			"CloseTime": data.CloseTime[i],
			"Open":      data.Open[i],
			"High":      data.High[i],
			"Low":       data.Low[i],
			"Close":     data.Close[i],
			// "MssBearish": data.MssBearish[i],
			// "MssBullish": data.MssBullish[i],
			"BullishLP": data.BullishLP[i],
			"BearishLP": data.BearishLP[i],
			// "BearishOTE": data.BearishOTE[i],
			// "BullishOTE": data.BullishOTE[i],
			"SwingHighs": data.SwingHighs[i],
			"SwingLows":  data.SwingLows[i],
			"STOPLOSS":   data.StopLoss[i],
			"BUY":        data.BUY[i],
			"SELL":       data.SELL[i],
		}
		results[i] = row
	}

	return results
}
