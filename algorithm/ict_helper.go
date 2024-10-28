package algorithm

import (
	"math"
	"trading/common"
)

func GetOptimalTradeEntry(candles []common.Candle, period, entryPeriod, trendConstant int) ([]int, []int) {
	n := len(candles)
	bullishOTE := make([]int, n)
	bearishOTE := make([]int, n)

	// 1. get swing highs and swing lows
	// 2. map out range of swing highs and swing lows
	// 3. for each pair, map out the fibbonacci range
	// 4. (bullish) check if
	// 	  (bearish) check if the price moves up into the 0.62 - 0.79 zone (touches 0.705) and moves down again
	// 		- if yes, entry point

	swingHighs, swingLows := GetSwingHighLows(candles)
	i := 0
	for i < n {
		if swingHighs[i] > 0 {
			j := i + 1
			for j < n && j < i+period {
				if swingLows[j] > 0 {
					middle := getFibbonacciRange(swingHighs[i], swingLows[j])
					entryIdx := findEntry(candles, middle, j+1, j+1+entryPeriod, trendConstant, false)
					if entryIdx > 0 {
						fillRange(bearishOTE, i, entryIdx)
						bearishOTE[entryIdx] = 2
						i = entryIdx
					}
				}
				j++
			}
		}

		if swingLows[i] > 0 {
			j := i + 1
			for j < n && j < i+period {
				if swingHighs[j] > 0 {
					middle := getFibbonacciRange(swingLows[i], swingHighs[j])
					entryIdx := findEntry(candles, middle, j+1, j+1+entryPeriod, trendConstant, true)
					if entryIdx > 0 {
						fillRange(bullishOTE, i, entryIdx)
						bullishOTE[entryIdx] = 2
						i = entryIdx
					}
				}
				j++
			}
		}
		i++
	}

	return bullishOTE, bearishOTE
}

// bullish
func findEntry(candles []common.Candle, middle float64, start, end, trendConstant int, findUp bool) int {
	if start >= len(candles) || start >= end {
		return -1
	}

	idx := -1

	for i := start; i <= end && i < len(candles); i++ {
		if candles[i].Close > middle {
			idx = i
			break
		}
	}

	if idx < 0 {
		return -1
	}

	// check if the trend is as expected after hit
	trendCounter := 0
	for i := idx; i <= end && i < len(candles); i++ {
		if findUp && candles[i].Close > candles[i].Open {
			trendCounter++
		}

		if !findUp && candles[i].Open > candles[i].Close {
			trendCounter++
		}
	}

	if trendCounter < trendConstant {
		return -1
	}

	return idx

}

func getFibbonacciRange(swingHigh, swingLow float64) float64 {
	//high := math.Abs(swingHigh-swingLow) * 0.79
	//low := math.Abs(swingHigh-swingLow) * 0.62
	middle := math.Abs(swingHigh-swingLow) * 0.79
	return middle
}

/*
========== UNICORN MODEL =========
(WIP, NOT IN USE, NEED FURTHER IMPROVEMENT)
*/
func UnicornModel(candles []common.Candle, bbConstant, uniConstant int) ([]int, []int, []float64) {
	n := len(candles)
	bearishUM := make([]int, n)
	bullishUM := make([]int, n)
	stopLoss := make([]float64, n)

	//bearishOB, bullishOB := GetOrderBlock(candles)
	swingHighs, swingLows := GetSwingHighLows(candles)

	i := 0
	for i < n {
		// find bullish unicorn model

		// 1. find the immediate swing high point, check if its a breaker block candidate (bearish OB - 2 candles comparison prereq only)
		// 2. within 20-30 candles area of the marked breaker block candidate:-
		// 		3a. find lower low (must come before higher high)
		// 		3b. find higher high (must come after lower low)
		// 		3c. the distance of lower low to higher high must cross the breaker block
		//      3d. after higher high, the trend must move down towards breaker block and move up again
		// 4. find fvg between lower low and higher high

		if swingHighs[i] > 0 && isBreakerBlock(candles[i-1], candles[i], false) {
			k := i
			for k < i+bbConstant && k < n {
				lowerLowIdx := getLLorHH(i, k-1, swingLows, candles[i].Low, true) // 3a
				if lowerLowIdx > 0 && swingHighs[k] > swingHighs[i] {             // 3b
					if crossBBAndFoundFVG(candles, i, lowerLowIdx, k, true, true) { // 3c
						entry := testMovementBullish(k, k+bbConstant, candles[i], candles, uniConstant)
						if entry > 0 {
							bullishUM[entry] = 1
							i = k
							break
						}
					}
				}
				k++
			}
		}

		// find bearish unicorn model

		// 1. find the immediate swing low point, check if its a breaker block candidate (bullish OB - 2 candles comparison prereq only)
		// 2. within 20-30 candles area of the marked breaker block candidate:-
		// 		3a. find higher high (must come before lower low)
		// 		3b. find lower low (must come after higher high)
		// 		3c. the distance of higher high to lower low must cross the breaker block
		//      3d. after lower low, the trend must move up towards breaker block and move down again
		// 4. find fvg between higher high and lower low

		if swingLows[i] > 0 && isBreakerBlock(candles[i-1], candles[i], true) {
			k := i
			for k < i+bbConstant && k < n {
				higherHighIdx := getLLorHH(i, k-1, swingHighs, candles[i].High, false) // 3a
				if higherHighIdx > 0 && swingLows[i] > swingLows[k] {                  // 3b
					if crossBBAndFoundFVG(candles, i, higherHighIdx, k, false, false) { // 3c
						entry := testMovementBearish(k, k+bbConstant, candles[i], candles, uniConstant)
						if entry > 0 {
							bearishUM[entry] = 1
							i = k
							break
						}
					}
				}
				k++
			}
		}
		i++
	}

	return bearishUM, bullishUM, stopLoss
}

/*
Check if the range of (swinglow - swinghigh / swingHigh - swingLow) fulfills 2 conditions:
- crosses the breaker block area
- exists an FVG between the range
*/
func crossBBAndFoundFVG(candles []common.Candle, bbIdx, startIdx, endIdx int, findBullish, checkCrossUp bool) bool {
	// check if cross breaker block
	if checkCrossUp && (candles[startIdx].Low > candles[bbIdx].Low || candles[endIdx].High < candles[bbIdx].High) {
		return false
	}

	if !checkCrossUp && (candles[endIdx].Low > candles[bbIdx].Low || candles[startIdx].High < candles[bbIdx].High) {
		return false
	}

	// check if found any fvg in between
	for i := startIdx + 1; i <= endIdx-1; i++ {
		if isFVG(candles, i-1, i, i+1, findBullish) && CheckOverlap(candles[i], candles[bbIdx]) {
			return true
		}
	}

	return false
}

/*
Checks if there exists a bearish movement (trend going up, hit the breaker area, and trend goind down) between the range.
Returns the index where the trend changes.
*/
func testMovementBearish(start, end int, breakerArea common.Candle, candles []common.Candle, uniConstant int) int {
	if start >= end || end >= len(candles) {
		return -1
	}

	entry := -1
	needTrendUp := true

	prev := candles[start]
	upTrendCandles := []common.Candle{}
	downTrendCandles := []common.Candle{}

	for i := start + 1; i < end; i++ {
		// keep going up for now; ok
		if needTrendUp {
			if candles[i].Close >= breakerArea.Low && candles[i].Close <= breakerArea.High {
				// found point of turn
				needTrendUp = false
			}

			if candles[i].Close > prev.Close {
				upTrendCandles = append(upTrendCandles, candles[i])
			}

		} else {
			// check unicorn uptrend
			if entry < 0 && candles[i].Close > prev.Close {
				entry = i
			}

			if prev.Close > candles[i].Close {
				downTrendCandles = append(downTrendCandles, candles[i])
			}
		}

		prev = candles[i]
	}

	if len(upTrendCandles) >= uniConstant && len(downTrendCandles) >= uniConstant && entry > 0 {
		return entry
	}
	return -1
}

/*
Checks if there exists a bullish movement (trend going down, hit the breaker area, and trend goind down) between the range.
Returns the index where the trend changes.
*/
func testMovementBullish(start, end int, breakerArea common.Candle, candles []common.Candle, uniConstant int) int {
	if start >= end || end >= len(candles) {
		return -1
	}

	entry := -1
	needTrendDown := true

	prev := candles[start]
	downTrendCandles := []common.Candle{}
	upTrendCandles := []common.Candle{}

	for i := start + 1; i < end; i++ {
		// keep going down for now; ok
		if needTrendDown {
			if candles[i].Close >= breakerArea.Low && candles[i].Close <= breakerArea.High {
				// found point of turn
				needTrendDown = false
			}

			if prev.Close > candles[i].Close {
				downTrendCandles = append(downTrendCandles, candles[i])
			}

		} else {
			// check unicorn uptrend
			if entry < 0 && candles[i].Close > prev.Close {
				entry = i
			}

			if candles[i].Close > prev.Close {
				upTrendCandles = append(upTrendCandles, candles[i])
			}
		}

		prev = candles[i]
	}

	if len(downTrendCandles) >= uniConstant && len(upTrendCandles) >= uniConstant && entry > 0 {
		return entry
	}
	return -1
}

// checks if the candle is a breaker block (with prev candle's stats).
func isBreakerBlock(prev common.Candle, curr common.Candle, checkBullish bool) bool {
	if checkBullish {
		return !isBullish(prev) && isBullish(curr) && prev.Low > curr.Low && curr.Close > prev.High
	} else {
		return isBullish(prev) && !isBullish(curr) && curr.High > prev.High && prev.Low > curr.Close
	}
}

// gets the index of Lower Low or Higher High from a range (swing lows or swing highs).
func getLLorHH(start, end int, swing []float64, target float64, findLL bool) int {
	if start >= end || end >= len(swing) {
		return -1
	}

	found := -1.0
	idx := -1

	if findLL {
		found = math.Inf(1)
		for i := start; i <= end; i++ {
			if swing[i] != 0 && swing[i] < target && swing[i] < found {
				found = swing[i]
				idx = i
			}
		}
	} else {
		found = math.Inf(-1)
		for i := start; i <= end; i++ {
			if swing[i] != 0 && swing[i] > target && swing[i] > found {
				found = swing[i]
				idx = i
			}
		}
	}

	if found != math.Inf(1) {
		return idx
	}

	return -1
}

// checks if 2 candles overlaps.
func CheckOverlap(candle1, candle2 common.Candle) bool {
	// Ensure the high and low are in correct order
	high1, low1 := max(candle1.High, candle1.Close), min(candle1.Low, candle1.Close)
	high2, low2 := max(candle2.High, candle2.Close), min(candle2.Low, candle2.Close)

	// Check if the ranges overlap
	return high1 >= low2 && high2 >= low1
}

/*
========== BREAKER BLOCK STRATEGY =========
(WIP, NOT IN USE, NEED FURTHER IMPROVEMENT)
*/
func GetBreakerBlock(candle []common.Candle, constant int) ([]int, []int) {
	// find H, L, HH, LL (bearish)
	// find L, H, LL, HH (bullish)
	n := len(candle)
	candleHighs := make([]float64, n)
	candleLows := make([]float64, n)

	for i := 1; i < n-1; i++ {
		if candle[i-1].Low > candle[i].Low && candle[i+1].Low > candle[i].Low {
			candleLows[i] = candle[i].Low
		}

		if candle[i-1].High < candle[i].High && candle[i+1].High > candle[i].High {
			candleHighs[i] = candle[i].High
		}
	}

	// find bb area
	bullishBB := make([]int, n)
	bearishBB := make([]int, n)

	i := 0
	for i < n {
		if candleLows[i] > 0 {
			ok, l := testBullishBB(i, n, constant, candleHighs, candleLows)
			if ok {
				fillRange(bullishBB, i+1, l)
			}
		} else if candleHighs[i] > 0 {
			ok, l := testBearishBB(i, n, constant, candleHighs, candleLows)
			if ok {
				fillRange(bearishBB, i+1, l)
			}
		}

	}

	return bearishBB, bullishBB
}

// checks if a series of low, higher high and lower low exists in a range. Returns the index of the lower low.
func testBearishBB(i, n, constant int, candleHighs, candleLows []float64) (bool, int) {
	j := i
	low := -1.00
	lowerLow := -1.00
	higherHigh := -1.00
	// find low
	for j < i+constant && j < n {
		if candleLows[j] > 0 && candleLows[j] < candleHighs[i] {
			low = candleHighs[j]
			break
		}
		j++
	}

	// find higher high
	k := j
	for k < i+j+constant && k < n {
		if candleHighs[k] > 0 && candleHighs[k] > candleHighs[i] {
			higherHigh = candleHighs[k]
			break
		}
		k++
	}

	// find lower low
	l := k
	for l < i+j+k+constant && l < n {
		if candleLows[l] > 0 && low > candleLows[l] {
			lowerLow = candleLows[l]
			break
		}
		l++
	}

	return low > 0 && lowerLow > 0 && higherHigh > 0, l
}

// checks if a series of high, lower low and higher high exists in a range. Returns the index of the higher high.
func testBullishBB(i, n, constant int, candleHighs, candleLows []float64) (bool, int) {
	j := i
	high := -1.00
	lowerLow := -1.00
	higherHigh := -1.00
	// find high
	for j < i+constant && j < n {
		if candleHighs[j] > 0 && candleHighs[j] > candleLows[i] {
			high = candleHighs[j]
			break
		}
		j++
	}

	// find lower low
	k := j
	for k < i+j+constant && k < n {
		if candleLows[k] > 0 && candleLows[i] > candleLows[k] {
			lowerLow = candleLows[k]
			break
		}
		k++
	}

	// find higher high
	l := k
	for l < i+j+k+constant && l < n {
		if candleHighs[l] > 0 && candleHighs[l] > high {
			higherHigh = candleHighs[l]
			break
		}
		l++
	}

	return high > 0 && lowerLow > 0 && higherHigh > 0, l
}

/* ================ MISCELLANEOUS FUNCTIONS ===========
These functions are written independently (they do not depend on other functions, but other functions might depend on them based on definition.)
These functions are currently not used. (WIP)
*/

/*
Calculate FVGs from candles using High, Low, Open, Close. Return array with len(candle). Bearish array, Bullish array.

CONDITION:
Bearish: 3 candles (constant decreasing) - gap between 1st candle low and 3rd candle high

Bullish: 3 candles (constant increasing) - gap between 1st candle high and 3rd candle low

If exists a gap at index 2 of candles with length 5, return [false, false, true, false, false] instead of [true].
*/
func GetFairValueGaps(candles []common.Candle) ([]int, []int) { // bearish array, bullish array
	n := len(candles)
	bearish := make([]int, n)
	bullish := make([]int, n)

	// Ensure we have at least 3 candles to check for FVGs
	if n < 3 {
		return bearish, bullish
	}

	// Loop through candles to find FVGs
	for i := 1; i < n-1; i++ {
		prev := candles[i-1]
		curr := candles[i]
		next := candles[i+1]

		if math.Abs(curr.Close-curr.Open) >= 0.7 && // body is wide enough
			math.Abs(curr.Close-curr.Open) > math.Abs(prev.Close-prev.Open) && math.Abs(curr.Close-curr.Open) > math.Abs(next.Close-next.Open) { // middle has the biggest body

			if prev.High > curr.High && curr.High > next.High && // decreasing
				prev.Low < math.Max(curr.Open, curr.Close) && next.High > math.Min(curr.Open, curr.Close) && // wicks fall between body
				next.High < prev.Low { // no overlaps
				bearish[i] = 1
			}

			if prev.High < curr.High && curr.High < next.High && // increasing
				prev.High > math.Min(curr.Open, curr.Close) && next.Low < math.Max(curr.Open, curr.Close) && // wicks fall between body
				prev.High < next.Low { // no overlaps
				// Check for a gap between the High of the 1st candle and the Low of the 3rd candle
				bullish[i] = 1
			}
		}
	}
	return bearish, bullish
}

/*
Calculate OBs from candles using High, Low, Open, Close. Return array with len(candle). Bearish OB, Bullish OB.

(bullish candle: close > open; bearish candle: close < open)

CONDITION:
Bearish OB: 1st candle should be bullish; 2nd candle bearish; price should be below low of the bearish candle; 2nd candle's close > 1st candle's bearish

Bullish OB: 1st candle should be bearish; 2nd candle bullish; price should be below low of the bearish candle; 2nd candle's close > 1st candle's bearish

Source: https://innercircletrader.net/tutorials/ict-order-block/

If exists an OB at index 0 of candles with length 3, return [true, false, false] instead of [true].
*/
func GetOrderBlock(candles []common.Candle) ([]int, []int) { // bearish OB array, bullish OB array
	n := len(candles)
	bearishOB := make([]int, n)
	bullishOB := make([]int, n)

	if n < 2 {
		return bearishOB, bullishOB
	}

	for i := 1; i < n; i++ {

		if candles[i-1].Close < candles[i-1].Open && // 1st candle is bearish
			candles[i].Close > candles[i].Open && // 2nd candle is bullish
			candles[i].Low < candles[i-1].Low && // 2nd candle grabs the low of 1st candle
			candles[i].Close > candles[i-1].High { // 2nd candle closes above the high of 1st candle
			bullishOB[i] = 1
		}

		if candles[i-1].Close > candles[i-1].Open && // 1st candle is bullish
			candles[i].Close < candles[i].Open && // 2nd candle is bearish
			candles[i].High > candles[i-1].High && // 2nd candle grabs the high of 1st candle
			candles[i].Close < candles[i-1].Low { // 2nd candle closes below the low of 1st candle
			bearishOB[i] = 1
		}
	}

	return bearishOB, bullishOB
}

/*
Calculate both buy stops and sell stops by finding previous highs/lows using candles' high/lows.

If exist a previous high for multiple candles, those indices share the same buy-stop price. Same for sell-stops.

Source: https://innercircletrading.website/ict-liquidity-trading-strategy/
*/
func GetLiquidity(candles []common.Candle, constant int) ([]float64, []float64) {
	n := len(candles)
	buyStops := make([]float64, n)
	sellStops := make([]float64, n)

	if len(candles) <= 0 {
		return buyStops, sellStops
	}

	prevH := candles[0].High
	prevL := candles[0].Low

	constantCounter := 0
	for i := 0; i < n; i++ {
		if constantCounter == constant {
			constantCounter = 0
			prevH = candles[i].High
			prevL = candles[i].Low
		}

		buyStops[i] = math.Max(prevH, candles[i].High)
		prevH = math.Max(prevH, candles[i].High)

		sellStops[i] = math.Min(prevL, candles[i].Low)
		prevL = math.Min(prevL, candles[i].Low)
		constantCounter++
	}

	return buyStops, sellStops
}

// checks if the candle has a huge body.
func HasBigBody(candle common.Candle, sizeConstant float64) bool {
	return math.Abs(candle.Close-candle.Open) >= sizeConstant
}
