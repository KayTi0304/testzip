package algorithm

// import (
// 	"math"
// 	"trading/common"

// 	"github.com/adshao/go-binance/v2/futures"
// )

// /*
// Calculate FVGs from candles using High, Low, Open, Close. Return array with len(candle). Bearish array, Bullish array.

// CONDITION:
// Bearish: 3 candles (constant decreasing) - gap between 1st candle low and 3rd candle high

// Bullish: 3 candles (constant increasing) - gap between 1st candle high and 3rd candle low

// If exists a gap at index 2 of candles with length 5, return [false, false, true, false, false] instead of [true].
// */
// func GetFairValueGaps(candles []common.Candle) ([]int, []int) { // bearish array, bullish array
// 	n := len(candles)
// 	bearish := make([]int, n)
// 	bullish := make([]int, n)

// 	// Ensure we have at least 3 candles to check for FVGs
// 	if n < 3 {
// 		return bearish, bullish
// 	}

// 	// Loop through candles to find FVGs
// 	for i := 1; i < n-1; i++ {
// 		prev := candles[i-1]
// 		curr := candles[i]
// 		next := candles[i+1]

// 		if math.Abs(curr.Close-curr.Open) >= 0.7 && // body is wide enough
// 			math.Abs(curr.Close-curr.Open) > math.Abs(prev.Close-prev.Open) && math.Abs(curr.Close-curr.Open) > math.Abs(next.Close-next.Open) { // middle has the biggest body

// 			if prev.High >= curr.High && curr.High >= next.High && // decreasing
// 				prev.Low <= math.Max(curr.Open, curr.Close) && next.High >= math.Min(curr.Open, curr.Close) && // wicks fall between body
// 				next.High < prev.Low { // no overlaps
// 				bearish[i] = 1
// 			}

// 			if prev.High <= curr.High && curr.High <= next.High && // increasing
// 				prev.High >= math.Min(curr.Open, curr.Close) && next.Low <= math.Max(curr.Open, curr.Close) && // wicks fall between body
// 				prev.High < next.Low { // no overlaps
// 				// Check for a gap between the High of the 1st candle and the Low of the 3rd candle
// 				bullish[i] = 1
// 			}
// 		}
// 	}
// 	return bearish, bullish
// }

// /*
// Calculate OBs from candles using High, Low, Open, Close. Return array with len(candle). Bearish OB, Bullish OB.

// (bullish candle: close > open; bearish candle: close < open)

// CONDITION:
// Bearish OB: 1st candle should be bullish; 2nd candle bearish; price should be below low of the bearish candle; 2nd candle's close > 1st candle's bearish

// Bullish OB: 1st candle should be bearish; 2nd candle bullish; price should be below low of the bearish candle; 2nd candle's close > 1st candle's bearish

// Source: https://innercircletrader.net/tutorials/ict-order-block/

// If exists an OB at index 0 of candles with length 3, return [true, false, false] instead of [true].
// */
// func GetOrderBlock(candles []common.Candle) ([]int, []int) { // bearish OB array, bullish OB array
// 	n := len(candles)
// 	bearishOB := make([]int, n)
// 	bullishOB := make([]int, n)

// 	if n < 2 {
// 		return bearishOB, bullishOB
// 	}

// 	for i := 1; i < n; i++ {

// 		if candles[i-1].Close < candles[i-1].Open && // 1st candle is bearish
// 			candles[i].Close > candles[i].Open && // 2nd candle is bullish
// 			candles[i].Low < candles[i-1].Low && // 2nd candle grabs the low of 1st candle
// 			candles[i].Close > candles[i-1].High { // 2nd candle closes above the high of 1st candle
// 			bullishOB[i] = 1
// 		}

// 		if candles[i-1].Close > candles[i-1].Open && // 1st candle is bullish
// 			candles[i].Close < candles[i].Open && // 2nd candle is bearish
// 			candles[i].High > candles[i-1].High && // 2nd candle grabs the high of 1st candle
// 			candles[i].Close < candles[i-1].Low { // 2nd candle closes below the low of 1st candle
// 			bearishOB[i] = 1
// 		}
// 	}

// 	return bearishOB, bullishOB
// }

// /*
// Calculate both buy stops and sell stops by finding previous highs/lows using candles' high/lows.

// If exist a previous high for multiple candles, those indices share the same buy-stop price. Same for sell-stops.

// Source: https://innercircletrading.website/ict-liquidity-trading-strategy/
// */
// func GetLiquidity(candles []common.Candle, constant int) ([]float64, []float64) {
// 	n := len(candles)
// 	buyStops := make([]float64, n)
// 	sellStops := make([]float64, n)

// 	if len(candles) <= 0 {
// 		return buyStops, sellStops
// 	}

// 	prevH := candles[0].High
// 	prevL := candles[0].Low

// 	constantCounter := 0
// 	for i := 0; i < n; i++ {
// 		if constantCounter == constant {
// 			constantCounter = 0
// 			prevH = candles[i].High
// 			prevL = candles[i].Low
// 		}

// 		buyStops[i] = math.Max(prevH, candles[i].High)
// 		prevH = math.Max(prevH, candles[i].High)

// 		sellStops[i] = math.Min(prevL, candles[i].Low)
// 		prevL = math.Min(prevL, candles[i].Low)
// 		constantCounter++
// 	}

// 	return buyStops, sellStops
// }

// /*
// Calculate MSSs from candles using High, Low, Open Close. This function WILL depend on GetFairValueGaps to determine displacement.

// CONDITION:
// Bullish MSS: displacement's movement breaks the previous swing high and is bullish displacement

// Bearish MSS: displacement's movement breaks the previous swing low and is bearish displacement

// Source: https://innercircletrader.net/tutorials/ict-market-structure-shift/

// 	https://fxopen.com/blog/en/market-structure-shift-meaning-and-use-in-ict-trading/#:~:text=A%20Market%20Structure%20Shift%20(MSS,through%20a%20key%20market%20level.

// The MSS is an area; the return array should indicate the area of the MSS. For example, if the start of the MSS is at index 1 and ends at index 3,
// the array should look like [false, true, true, true, ...].
//   - start of MSS: the low point of the swing high/low
//   - end of MSS: when the displacement first breaks the low/high of swing high/low
// */
// func GetMarketStructureShift(candles []common.Candle, constant, displacementCounter int) ([]int, []int, []float64, []float64) { // bearish, bullish
// 	//TODO:
// 	// 1. Find swing high (3 candles; middle candle has highest high)
// 	// 2. Find swing low (3 candles: middle candle has lowest high)
// 	swingHighs, swingLows := GetSwingHighLows(candles)

// 	// 3. Get FVG (tbh not necessary also; just need obvious downward/upward trend)
// 	//fvgBearish, fvgBullish := GetFairValueGaps(candles)

// 	// 4. Go through each candle, find the MSS area (update the previous highest/lowest point with swing high/low if needed)
// 	n := len(candles)
// 	mssBullish := make([]int, n)
// 	mssBearish := make([]int, n)

// 	i := 0
// 	for i < n {
// 		if swingHighs[i] > 0 {
// 			j := i
// 			for j < i+constant {
// 				ok, start, breakHigh := isBullishDisplacement(candles, i+1, j, displacementCounter)
// 				if ok && achievedLLInBetween(i+1, start, swingLows) {
// 					// ignoring huge candles for displacement; using solely fvg as displacement
// 					fillRange(mssBullish, i, breakHigh)
// 					mssBullish[breakHigh] = 2
// 					j = breakHigh
// 					break
// 				}

// 				ok, start, breakHigh = IsBullishWithFVG(candles, i+1, j, displacementCounter)
// 				if ok && achievedLLInBetween(i+1, start, swingLows) {
// 					// ignoring huge candles for displacement; using solely fvg as displacement
// 					fillRange(mssBullish, i, breakHigh)
// 					mssBullish[breakHigh] = 2
// 					j = breakHigh
// 					break
// 				}

// 				j++
// 			}
// 			i = j + 1

// 		} else if swingLows[i] > 0 {
// 			j := i
// 			for j < i+constant {

// 				ok, start, breakLow := isBearishDisplacement(candles, i+1, j, displacementCounter)

// 				if ok && achievedHHInBetween(i+1, start, swingHighs) {
// 					// ignoring huge candles for displacement; using solely fvg as displacement
// 					fillRange(mssBearish, i, breakLow)
// 					mssBearish[breakLow] = 2
// 					j = breakLow
// 					break
// 				}

// 				ok, start, breakLow = IsBearishWithFVG(candles, i+1, j, displacementCounter)
// 				if ok && achievedHHInBetween(i+1, start, swingLows) {
// 					// ignoring huge candles for displacement; using solely fvg as displacement
// 					fillRange(mssBullish, i, breakLow)
// 					mssBullish[breakLow] = 2
// 					j = breakLow
// 					break
// 				}

// 				j++
// 			}
// 			i = j + 1

// 		} else {
// 			i++
// 		}
// 	}

// 	return mssBearish, mssBullish, swingHighs, swingLows
// }

// func achievedHHInBetween(start, end int, swingHighs []float64) bool {
// 	foundHH := false

// 	for i := start; i <= end; i++ {
// 		if swingHighs[i] > 0 {
// 			foundHH = true
// 			break
// 		}
// 	}

// 	return foundHH
// }

// func achievedLLInBetween(start, end int, swingLows []float64) bool {
// 	foundLL := false

// 	for i := start; i <= end; i++ {
// 		if swingLows[i] > 0 {
// 			foundLL = true
// 			break
// 		}
// 	}

// 	return foundLL
// }

// func isBearishDisplacement(candles []common.Candle, start, end, displacementCounter int) (bool, int, int) {
// 	if start >= end || end >= len(candles) {
// 		return false, -1, -1
// 	}

// 	// check if exist displacement between this range
// 	displacements := []int{}
// 	prevC := candles[start].Close
// 	for i := start; i <= end; i++ {
// 		if !isBullish(candles[i]) && candles[i].Close < prevC {
// 			displacements = append(displacements, i)
// 			prevC = candles[i].Close
// 		}
// 	}

// 	if len(displacements) < displacementCounter {
// 		return false, -1, -1
// 	}

// 	// find the break low
// 	endd := -1
// 	for _, idx := range displacements {
// 		if candles[idx].Close > candles[start-1].Low {
// 			endd = idx
// 		}
// 	}

// 	return endd > 0 && len(displacements) >= displacementCounter, displacements[0], endd
// }

// func isBullishDisplacement(candles []common.Candle, start, end, displacementCounter int) (bool, int, int) {
// 	if start >= end || end >= len(candles) {
// 		return false, -1, -1
// 	}

// 	// check if exist displacement between this range
// 	displacements := []int{}
// 	prevC := candles[start].Close
// 	for i := start; i <= end; i++ {
// 		if isBullish(candles[i]) && candles[i].Close > prevC {
// 			displacements = append(displacements, i)
// 			prevC = candles[i].Close
// 		}
// 	}

// 	if len(displacements) < displacementCounter {
// 		return false, -1, -1
// 	}

// 	// find the break high
// 	endd := -1
// 	for _, idx := range displacements {
// 		if candles[idx].Close > candles[start-1].High {
// 			endd = idx
// 		}
// 	}

// 	return endd > 0 && len(displacements) >= displacementCounter, displacements[0], endd
// }

// func IsBearishWithFVG(candles []common.Candle, start, end, fvgCounter int) (bool, int, int) {
// 	if start >= end || end >= len(candles) || end-start <= 2 {
// 		return false, -1, -1
// 	}

// 	// check if exist fvg between this range
// 	fvg := []int{}
// 	for i := start + 1; i <= end-1; i++ {
// 		if isFVG(candles, i-1, i, i+1, false) {
// 			fvg = append(fvg, i)
// 		}
// 	}

// 	if len(fvg) < fvgCounter {
// 		return false, -1, -1
// 	}

// 	// find the break high
// 	endd := -1
// 	for _, idx := range fvg {
// 		if candles[idx].Close > candles[start-1].High {
// 			endd = idx
// 		}
// 	}

// 	return endd > 0 && len(fvg) >= fvgCounter, fvg[0], endd
// }

// func IsBullishWithFVG(candles []common.Candle, start, end, fvgCounter int) (bool, int, int) {
// 	if start >= end || end >= len(candles) || end-start <= 2 {
// 		return false, -1, -1
// 	}

// 	// check if exist fvg between this range
// 	fvg := []int{}
// 	for i := start + 1; i <= end-1; i++ {
// 		if isFVG(candles, i-1, i, i+1, true) {
// 			fvg = append(fvg, i)
// 		}
// 	}

// 	if len(fvg) < fvgCounter {
// 		return false, -1, -1
// 	}

// 	// find the break high
// 	endd := -1
// 	for _, idx := range fvg {
// 		if candles[idx].Close > candles[start-1].High {
// 			endd = idx
// 		}
// 	}

// 	return endd > 0 && len(fvg) >= fvgCounter, fvg[0], endd
// }

// func isFVG(candles []common.Candle, i, j, k int, findBullish bool) bool {
// 	if findBullish {
// 		return candles[i].High <= candles[j].High && candles[j].High <= candles[k].High && // ascending
// 			candles[j].Open < candles[i].High && candles[j].Close > candles[k].Low && // within body
// 			candles[i].High < candles[k].Low // no overlap
// 	} else {
// 		return candles[i].High >= candles[j].High && candles[j].High >= candles[k].High && // descending
// 			candles[j].Open >= candles[i].Low && candles[j].Close <= candles[k].High && // within body
// 			candles[i].Low > candles[k].High // no overlap
// 	}
// }

// func HasBigBody(candles []common.Candle, idx int, sizeConstant float64) bool {
// 	return math.Abs(candles[idx].Close-candles[idx].Open) >= sizeConstant
// }

// func GetSwingHighLows(candles []common.Candle) ([]float64, []float64) {
// 	n := len(candles)
// 	swingHighs := make([]float64, n)
// 	swingLows := make([]float64, n)

// 	if n < 3 {
// 		return swingHighs, swingLows
// 	}

// 	for i := 1; i < n-1; i++ {
// 		if // isBullish(candles[i-1]) && isBullish(candles[i]) && !isBullish(candles[i+1]) &&
// 		candles[i].High > candles[i-1].High && candles[i].High >= candles[i+1].High {
// 			// swing high
// 			swingHighs[i] = candles[i].High
// 		}

// 		if // !isBullish(candles[i-1]) && !isBullish(candles[i]) && isBullish(candles[i+1]) &&
// 		candles[i].Low <= candles[i-1].Low && candles[i].Low <= candles[i+1].Low {
// 			// swing low
// 			swingLows[i] = candles[i].Low
// 		}
// 	}

// 	return swingHighs, swingLows
// }

// func isBullish(candle common.Candle) bool {
// 	return candle.Close > candle.Open
// }

// func fillRange(slice []int, start, end int) {
// 	for i := start; i <= end; i++ {
// 		slice[i] = 1
// 	}
// }

// func GetBreakerBlock(candle []common.Candle, constant int) ([]int, []int) {
// 	// find H, L, HH, LL (bearish)
// 	// find L, H, LL, HH (bullish)
// 	n := len(candle)
// 	candleHighs := make([]float64, n)
// 	candleLows := make([]float64, n)

// 	for i := 1; i < n-1; i++ {
// 		if candle[i-1].Low > candle[i].Low && candle[i+1].Low > candle[i].Low {
// 			candleLows[i] = candle[i].Low
// 		}

// 		if candle[i-1].High < candle[i].High && candle[i+1].High > candle[i].High {
// 			candleHighs[i] = candle[i].High
// 		}
// 	}

// 	// find bb area
// 	bullishBB := make([]int, n)
// 	bearishBB := make([]int, n)

// 	i := 0
// 	for i < n {
// 		if candleLows[i] > 0 {
// 			ok, l := testBullishBB(i, n, constant, candleHighs, candleLows)
// 			if ok {
// 				fillRange(bullishBB, i+1, l)
// 			}
// 		} else if candleHighs[i] > 0 {
// 			ok, l := testBearishBB(i, n, constant, candleHighs, candleLows)
// 			if ok {
// 				fillRange(bearishBB, i+1, l)
// 			}
// 		}

// 	}

// 	return bearishBB, bullishBB
// }

// func testBearishBB(i, n, constant int, candleHighs, candleLows []float64) (bool, int) {
// 	j := i
// 	low := -1.00
// 	lowerLow := -1.00
// 	higherHigh := -1.00
// 	// find low
// 	for j < i+constant && j < n {
// 		if candleLows[j] > 0 && candleLows[j] < candleHighs[i] {
// 			low = candleHighs[j]
// 			break
// 		}
// 		j++
// 	}

// 	// find higher high
// 	k := j
// 	for k < i+j+constant && k < n {
// 		if candleHighs[k] > 0 && candleHighs[k] > candleHighs[i] {
// 			higherHigh = candleHighs[k]
// 			break
// 		}
// 		k++
// 	}

// 	// find lower low
// 	l := k
// 	for l < i+j+k+constant && l < n {
// 		if candleLows[l] > 0 && low > candleLows[l] {
// 			lowerLow = candleLows[l]
// 			break
// 		}
// 		l++
// 	}

// 	return low > 0 && lowerLow > 0 && higherHigh > 0, l
// }

// func testBullishBB(i, n, constant int, candleHighs, candleLows []float64) (bool, int) {
// 	j := i
// 	high := -1.00
// 	lowerLow := -1.00
// 	higherHigh := -1.00
// 	// find high
// 	for j < i+constant && j < n {
// 		if candleHighs[j] > 0 && candleHighs[j] > candleLows[i] {
// 			high = candleHighs[j]
// 			break
// 		}
// 		j++
// 	}

// 	// find lower low
// 	k := j
// 	for k < i+j+constant && k < n {
// 		if candleLows[k] > 0 && candleLows[i] > candleLows[k] {
// 			lowerLow = candleLows[k]
// 			break
// 		}
// 		k++
// 	}

// 	// find higher high
// 	l := k
// 	for l < i+j+k+constant && l < n {
// 		if candleHighs[l] > 0 && candleHighs[l] > high {
// 			higherHigh = candleHighs[l]
// 			break
// 		}
// 		l++
// 	}

// 	return high > 0 && lowerLow > 0 && higherHigh > 0, l
// }

// /*
// return point to trade and stoploss (in array). (bullish unicorn, bearish unicorn, stoploss)
// */
// func UnicornModel(candles []common.Candle, constant, bbConstant, uniConstant int) ([]int, []int, []float64) {
// 	n := len(candles)
// 	bearishUM := make([]int, n)
// 	bullishUM := make([]int, n)
// 	stopLoss := make([]float64, n)

// 	//bearishOB, bullishOB := GetOrderBlock(candles)
// 	swingHighs, swingLows := GetSwingHighLows(candles)

// 	i := 0
// 	for i < n {
// 		// find bullish unicorn model

// 		// 1. if current point is a swing low, mark it
// 		// 2. find the immediate swing high point, check if its a breaker block candidate (bearish OB - 2 candles comparison prereq only)
// 		// 3. within 20-30 candles area of the marked breaker block candidate:-
// 		// 		3a. find lower low (must come before higher high)
// 		// 		3b. find higher high (must come after lower low)
// 		// 		3c. the distance of lower low to higher high must cross the breaker block
// 		//      3d. after higher high, the trend must move down towards breaker block and move up again
// 		// 4. find fvg between lower low and higher high
// 		if swingLows[i] > 0 {
// 			j := i
// 			for j < i+constant && j < n {
// 				if swingHighs[j] > 0 && isBreakerBlock(candles[j-1], candles[j], false) {
// 					k := j
// 					for k < j+bbConstant && k < n {
// 						lowerLowIdx := getLLorHH(j, k-1, swingLows, candles[i].Low, true) // 3a
// 						if lowerLowIdx > 0 && swingHighs[k] > swingHighs[j] {             // 3b
// 							if crossBBAndFoundFVG(candles, j, lowerLowIdx, k, true, true) { // 3c
// 								entry := testMovementBullish(k, k+bbConstant, candles[i], candles, uniConstant)
// 								if entry > 0 {
// 									bullishUM[entry] = 1
// 									i = k
// 									break
// 								}
// 							}
// 						}
// 						k++
// 					}
// 				}
// 				j++
// 			}
// 			// find bearish unicorn model

// 			// 1. if current point is a swing high, mark it
// 			// 2. find the immediate swing low point, check if its a breaker block candidate (bullish OB - 2 candles comparison prereq only)
// 			// 3. within 20-30 candles area of the marked breaker block candidate:-
// 			// 		3a. find higher high (must come before lower low)
// 			// 		3b. find lower low (must come after higher high)
// 			// 		3c. the distance of higher high to lower low must cross the breaker block
// 			//      3d. after lower low, the trend must move up towards breaker block and move down again
// 			// 4. find fvg between higher high and lower low
// 		} else if swingHighs[i] > 0 {
// 			j := i
// 			for j < i+constant && j < n {
// 				if swingLows[j] > 0 && isBreakerBlock(candles[j-1], candles[j], true) {
// 					k := j
// 					for k < j+bbConstant && k < n {
// 						higherHighIdx := getLLorHH(j, k-1, swingHighs, candles[i].High, false) // 3a
// 						if higherHighIdx > 0 && swingLows[k] > swingLows[j] {                  // 3b
// 							if crossBBAndFoundFVG(candles, j, higherHighIdx, k, false, false) { // 3c
// 								entry := testMovementBearish(k, k+bbConstant, candles[i], candles, uniConstant)
// 								if entry > 0 {
// 									bearishUM[entry] = 1
// 									i = k
// 									break
// 								}
// 							}
// 						}
// 						k++
// 					}
// 				}
// 				j++
// 			}
// 		}
// 		i++
// 	}

// 	return bearishUM, bullishUM, stopLoss
// }

// func crossBBAndFoundFVG(candles []common.Candle, bbIdx, startIdx, endIdx int, findBullish, checkCrossUp bool) bool {
// 	// check if cross breaker block
// 	if checkCrossUp && (candles[startIdx].Low > candles[bbIdx].Low || candles[endIdx].High < candles[bbIdx].High) {
// 		return false
// 	}

// 	if !checkCrossUp && (candles[endIdx].Low > candles[bbIdx].Low || candles[startIdx].High < candles[bbIdx].High) {
// 		return false
// 	}

// 	// check if found any fvg in between
// 	for i := startIdx + 1; i <= endIdx-1; i++ {
// 		if isFVG(candles, i-1, i, i+1, findBullish) && CheckOverlap(candles[i], candles[bbIdx]) {
// 			return true
// 		}
// 	}

// 	return false
// }

// func testMovementBearish(start, end int, breakerArea common.Candle, candles []common.Candle, uniConstant int) int {
// 	if start >= end || end >= len(candles) {
// 		return -1
// 	}

// 	entry := -1
// 	needTrendUp := true

// 	prev := candles[start]
// 	upTrendCandles := []common.Candle{}
// 	downTrendCandles := []common.Candle{}

// 	for i := start + 1; i < end; i++ {
// 		// keep going up for now; ok
// 		if needTrendUp {
// 			if candles[i].Close >= breakerArea.Low && candles[i].Close <= breakerArea.High {
// 				// found point of turn
// 				needTrendUp = false
// 			}

// 			if candles[i].Close > prev.Close {
// 				upTrendCandles = append(upTrendCandles, candles[i])
// 			}

// 		} else {
// 			// check unicorn uptrend
// 			if entry < 0 && candles[i].Close > prev.Close {
// 				entry = i
// 			}

// 			if prev.Close > candles[i].Close {
// 				downTrendCandles = append(downTrendCandles, candles[i])
// 			}
// 		}

// 		prev = candles[i]
// 	}

// 	if len(upTrendCandles) >= uniConstant && len(downTrendCandles) >= uniConstant && entry > 0 {
// 		return entry
// 	}
// 	return -1
// }

// func testMovementBullish(start, end int, breakerArea common.Candle, candles []common.Candle, uniConstant int) int {
// 	if start >= end || end >= len(candles) {
// 		return -1
// 	}

// 	entry := -1
// 	needTrendDown := true

// 	prev := candles[start]
// 	downTrendCandles := []common.Candle{}
// 	upTrendCandles := []common.Candle{}

// 	for i := start + 1; i < end; i++ {
// 		// keep going down for now; ok
// 		if needTrendDown {
// 			if candles[i].Close >= breakerArea.Low && candles[i].Close <= breakerArea.High {
// 				// found point of turn
// 				needTrendDown = false
// 			}

// 			if prev.Close > candles[i].Close {
// 				downTrendCandles = append(downTrendCandles, candles[i])
// 			}

// 		} else {
// 			// check unicorn uptrend
// 			if entry < 0 && candles[i].Close > prev.Close {
// 				entry = i
// 			}

// 			if candles[i].Close > prev.Close {
// 				upTrendCandles = append(upTrendCandles, candles[i])
// 			}
// 		}

// 		prev = candles[i]
// 	}

// 	if len(downTrendCandles) >= uniConstant && len(upTrendCandles) >= uniConstant && entry > 0 {
// 		return entry
// 	}
// 	return -1
// }

// func isBreakerBlock(prev common.Candle, curr common.Candle, checkBullish bool) bool {
// 	if checkBullish {
// 		return !isBullish(prev) && isBullish(curr) && prev.Low > curr.Low && curr.Close > prev.High
// 	} else {
// 		return isBullish(prev) && !isBullish(curr) && curr.High > prev.High && prev.Low > curr.Close
// 	}
// }

// func getLLorHH(start, end int, swing []float64, target float64, findLL bool) int {
// 	if start >= end || end >= len(swing) {
// 		return -1
// 	}

// 	found := -1.0
// 	idx := -1

// 	if findLL {
// 		found = math.Inf(1)
// 		for i := start; i <= end; i++ {
// 			if swing[i] != 0 && swing[i] < target && swing[i] < found {
// 				found = swing[i]
// 				idx = i
// 			}
// 		}
// 	} else {
// 		found = math.Inf(-1)
// 		for i := start; i <= end; i++ {
// 			if swing[i] != 0 && swing[i] > target && swing[i] > found {
// 				found = swing[i]
// 				idx = i
// 			}
// 		}
// 	}

// 	if found != math.Inf(1) {
// 		return idx
// 	}

// 	return -1
// }

// func CheckOverlap(candle1, candle2 common.Candle) bool {
// 	// Ensure the high and low are in correct order
// 	high1, low1 := max(candle1.High, candle1.Close), min(candle1.Low, candle1.Close)
// 	high2, low2 := max(candle2.High, candle2.Close), min(candle2.Low, candle2.Close)

// 	// Check if the ranges overlap
// 	return high1 >= low2 && high2 >= low1
// }

// type ICTData struct {
// 	OpenTime   []string
// 	CloseTime  []string
// 	Open       []float64
// 	High       []float64
// 	Low        []float64
// 	Close      []float64
// 	SwingHighs []float64
// 	SwingLows  []float64
// 	MssBullish []int
// 	MssBearish []int
// 	BUY        []float64
// 	SELL       []float64
// 	StopLoss   []float64
// }

// func HaThreeSupertrendsICT(klines []*futures.Kline, mssPeriod, sameTrendPeriod, stopLossPeriod, minimumTrendCounter, displacementCounter int) ICTData {
// 	candles := common.ExtractCandles(klines)
// 	length := len(candles)
// 	data := ICTData{
// 		OpenTime:   make([]string, length),
// 		CloseTime:  make([]string, length),
// 		Open:       make([]float64, length),
// 		High:       make([]float64, length),
// 		Low:        make([]float64, length),
// 		Close:      make([]float64, length),
// 		MssBullish: make([]int, length),
// 		MssBearish: make([]int, length),
// 		SwingHighs: make([]float64, length),
// 		SwingLows:  make([]float64, length),
// 		StopLoss:   make([]float64, length),
// 		BUY:        make([]float64, length),
// 		SELL:       make([]float64, length),
// 	}

// 	mssBearish, mssBullish, swingHighs, swingLows := GetMarketStructureShift(candles, mssPeriod, displacementCounter)
// 	//stopLossForSell, stopLossForBuy := GetLiquidity(candles, 30)
// 	bullishUnicorn, bearishUnicorn, _ := UnicornModel(candles, 10, 20, 10)

// 	currentAction := ""

// 	for i, candle := range candles {
// 		data.Open[i] = candle.Open
// 		data.Close[i] = candle.Close
// 		data.OpenTime[i] = candle.OpenTime
// 		data.CloseTime[i] = candle.CloseTime
// 		data.High[i] = candle.High
// 		data.Low[i] = candle.Low
// 		data.MssBearish[i] = mssBearish[i]
// 		data.MssBullish[i] = mssBullish[i]
// 		data.SwingHighs[i] = swingHighs[i]
// 		data.SwingLows[i] = swingLows[i]

// 		if currentAction == "SELL" {
// 			data.StopLoss[i] = candles[i].High + 0.3
// 		} else if currentAction == "BUY" {
// 			data.StopLoss[i] = candles[i].Low - 0.3
// 		}

// 		if mssBearish[i] == 2 {
// 			// find the start of this bearish mss and check if the trend before is upwards
// 			j := i - 1
// 			for j >= 0 && mssBearish[j] == 1 {
// 				j--
// 			}
// 			if isSameTrend(j, candles, sameTrendPeriod, minimumTrendCounter, true) {
// 				data.SELL[i] = candle.Close
// 				//data.StopLoss[i] = GetStopLoss(candles, i, stopLossPeriod, true)
// 				//data.StopLoss[i+1] = GetStopLoss(candles, i+1, stopLossPeriod, true)
// 				data.StopLoss[i] = candles[i].High + 0.3
// 				currentAction = "SELL"
// 			}
// 		}

// 		if mssBullish[i] == 2 {
// 			// find the start of this bearish mss and check if the trend before is downwards
// 			j := i - 1
// 			for j >= 0 && mssBullish[j] == 1 {
// 				j--
// 			}
// 			if isSameTrend(j, candles, sameTrendPeriod, minimumTrendCounter, false) {
// 				data.BUY[i] = candle.Close
// 				//data.StopLoss[i] = GetStopLoss(candles, i, stopLossPeriod, false)
// 				data.StopLoss[i] = candles[i].Low - 0.3
// 				currentAction = "BUY"
// 			}
// 		}

// 		if bearishUnicorn[i] > 0 {
// 			data.SELL[i] = candle.Close
// 			data.StopLoss[i] = candles[i].High + 0.3
// 		}

// 		if bullishUnicorn[i] > 0 {
// 			data.BUY[i] = candle.Close
// 			data.StopLoss[i] = candles[i].Low - 0.3
// 		}
// 	}
// 	common.SaveCsv4(data.ToMapSlice(), "mss")
// 	return data
// }

// func GetStopLoss(candles []common.Candle, idx, period int, findBuy bool) float64 {
// 	// see forward
// 	res := 0.0
// 	if findBuy {
// 		res = candles[idx].High
// 	} else {
// 		res = candles[idx].Low
// 	}

// 	for i := idx; i >= 0 && i >= idx-period; i-- {
// 		if findBuy {
// 			res = math.Max(res, candles[i].High)
// 		} else {
// 			res = math.Min(res, candles[i].Low)
// 		}
// 	}

// 	// see backward
// 	for i := idx; i < len(candles) && i <= idx-period; i++ {
// 		if findBuy {
// 			res = math.Max(res, candles[i].High)
// 		} else {
// 			res = math.Min(res, candles[i].Low)
// 		}
// 	}

// 	return res
// }

// func isSameTrend(idx int, candles []common.Candle, period, minimumTrendCounter int, detectUptrend bool) bool {
// 	if idx-period < 0 {
// 		return false
// 	}

// 	trendCounter := 0
// 	prevC := candles[idx].Close
// 	if detectUptrend {
// 		for i := idx - 1; i > idx-period; i-- {
// 			if prevC >= candles[i].Close { // doing in reverse
// 				trendCounter++
// 				prevC = candles[i].Close

// 			}
// 		}
// 	} else {
// 		for i := idx - 1; i > idx-period; i-- {
// 			if prevC <= candles[i].Close { // doing in reverse
// 				trendCounter++
// 				prevC = candles[i].Close
// 			}

// 		}
// 	}

// 	return trendCounter >= minimumTrendCounter
// }

// func (data *ICTData) ToMapSlice() []map[string]interface{} {
// 	length := len(data.Open) // Assuming all slices are the same length
// 	results := make([]map[string]interface{}, length)

// 	for i := 0; i < length; i++ {
// 		row := map[string]interface{}{
// 			"OpenTime":   data.OpenTime[i],
// 			"CloseTime":  data.CloseTime[i],
// 			"Open":       data.Open[i],
// 			"High":       data.High[i],
// 			"Low":        data.Low[i],
// 			"Close":      data.Close[i],
// 			"MssBearish": data.MssBearish[i],
// 			"MssBullish": data.MssBullish[i],
// 			"SwingHighs": data.SwingHighs[i],
// 			"SwingLows":  data.SwingLows[i],
// 			"STOPLOSS":   data.StopLoss[i],
// 			"BUY":        data.BUY[i],
// 			"SELL":       data.SELL[i],
// 		}
// 		results[i] = row
// 	}

// 	return results
// }
//
