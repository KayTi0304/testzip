package backtest

import (
	"math"
	"trading/algorithm"
	"trading/common"
)

func calPNL(margin, entryPrice, exitPrice, usdtValue float64) float64 {
	if usdtValue < 5 {
		usdtValue = 5
	}
	if entryPrice == 0 {
		return 0.0
	}
	buyAmt := usdtValue * margin
	quantity := buyAmt / entryPrice
	pnl := (exitPrice - entryPrice) * quantity

	return pnl
}

func CalLastPriceFrmROE(entry, cost float64, leverage int, inBuy bool, roe float64) float64 {
	positionMultiplier := 1.0
	if !inBuy {
		positionMultiplier = -1.0
	}
	// Calculate the amount of profit/loss based on the ROE and cost
	roeProfitLoss := (roe / 100) * cost
	// log.Printf("rpl: %.1f", roeProfitLoss)
	// Calculate the last price based on the entry price and the calculated profit/loss
	lastPrice := entry + ((roeProfitLoss / (cost * float64(leverage)) * entry) * positionMultiplier)
	// return math.Round(lastPrice)
	return lastPrice
}

func Backtestv2(data algorithm.ICTData, margin float64, tpPercent float64, investAmt int, investPercent float64, cutLossMultiplier float64) (common.TradeResult, common.SummaryMetrics) {
	results := common.TradeResult{
		OpenTime: data.OpenTime,
		Action:   make([]string, len(data.OpenTime)),
		Open:     data.Open,
		High:     data.High,
		Low:      data.Low,
		Close:    data.Close,
		// HaOpen:         data.HaOpen,
		// HaHigh:         data.HaHigh,
		// HaLow:          data.HaLow,
		// HaClose:        data.HaClose,
		Entry:          make([]float64, len(data.OpenTime)),
		Exit:           make([]float64, len(data.OpenTime)),
		TakeProfit:     make([]float64, len(data.OpenTime)),
		StopLoss:       data.StopLoss,
		WinLose:        make([]float64, len(data.OpenTime)),
		PNL:            make([]float64, len(data.OpenTime)),
		Balance:        make([]float64, len(data.OpenTime)),
		ClosePosReason: make([]string, len(data.OpenTime)),
	}
	var totalPNL, wallet float64
	var wins, losses, totalTrades int

	var inBuySell bool
	var currentAction string
	var currentEntry, currentTakeProfit float64

	wallet = float64(investAmt)
	investValue := float64(investAmt) * investPercent

	for i := 30; i < len(data.OpenTime); i++ {
		if data.Close[i] <= 0 || wallet <= 0 {
			continue
		}

		if inBuySell {
			takeProfitCondition := (currentAction == "BUY" && (data.Close[i] >= currentTakeProfit || data.High[i] >= currentTakeProfit)) ||
				(currentAction == "SELL" && (data.Close[i] <= currentTakeProfit || data.Low[i] <= currentTakeProfit))

			stopLossCondition := (currentAction == "BUY" && (data.Close[i] <= data.StopLoss[i] || data.Low[i] <= data.StopLoss[i])) ||
				(currentAction == "SELL" && (data.Close[i] >= data.StopLoss[i] || data.High[i] >= data.StopLoss[i]))

			// currentCloseRoe := calPNL(margin, currentEntry, data.Close[i], investValue)
			// currentHgLwRoe := calPNL(margin, currentEntry, data.Low[i], investValue)
			// if currentAction == "SELL" {
			// 	currentCloseRoe = calPNL(margin, data.Close[i], currentEntry, investValue)
			// 	currentHgLwRoe = calPNL(margin, data.High[i], currentEntry, investValue)
			// }

			if takeProfitCondition || stopLossCondition {
				exitPrice := currentTakeProfit
				closeReason := "hit TP"
				if stopLossCondition {
					exitPrice = data.StopLoss[i]
					closeReason = "hit SL"
				}
				pnl := calPNL(margin, currentEntry, exitPrice, investValue)
				if currentAction == "SELL" {
					pnl = calPNL(margin, exitPrice, currentEntry, investValue)
				}
				totalPNL += pnl
				wallet += pnl
				totalTrades++
				if pnl > 0 {
					wins++
					results.WinLose[i] = 1
				} else {
					losses++
					results.WinLose[i] = -1
				}

				results.Action[i] = currentAction
				results.Entry[i] = currentEntry
				results.Exit[i] = exitPrice
				results.TakeProfit[i] = currentTakeProfit
				results.StopLoss[i] = data.StopLoss[i]
				results.PNL[i] = pnl
				results.Balance[i] = wallet
				results.ClosePosReason[i] = closeReason
				inBuySell = false
			} else if (currentAction == "BUY" && data.SELL[i] != 0) || (currentAction == "SELL" && data.BUY[i] != 0) {
				exitPrice := data.Close[i]
				closeReason := "change trend"
				stopLossCondition := (currentAction == "BUY" && (data.Close[i] <= data.StopLoss[i-1] || data.Low[i] <= data.StopLoss[i-1])) ||
					(currentAction == "SELL" && (data.Close[i] >= data.StopLoss[i-1] || data.High[i] >= data.StopLoss[i-1]))
				if stopLossCondition {
					exitPrice = data.StopLoss[i-1]
					closeReason = "hit SL"
				}

				pnl := calPNL(margin, currentEntry, exitPrice, investValue)
				if currentAction == "SELL" {
					pnl = calPNL(margin, exitPrice, currentEntry, investValue)
				}
				totalPNL += pnl
				wallet += pnl
				totalTrades++
				if pnl > 0 {
					wins++
					results.WinLose[i] = 1
				} else {
					losses++
					results.WinLose[i] = -1
				}

				results.Action[i] = currentAction
				results.Entry[i] = currentEntry
				results.Exit[i] = exitPrice
				results.TakeProfit[i] = currentTakeProfit
				results.StopLoss[i] = data.StopLoss[i]
				results.PNL[i] = pnl
				results.Balance[i] = wallet
				results.ClosePosReason[i] = closeReason

				// currentAction = "SELL"
				if currentAction == "BUY" {
					currentAction = "SELL"
				} else {
					currentAction = "BUY"
				}
				currentEntry = data.Close[i]
				currentTakeProfit = CalLastPriceFrmROE(currentEntry, investValue, int(margin), currentAction == "BUY", tpPercent)
				// } else if (cutLossMultiplier > 0.0 && currentCloseRoe < 0 && math.Abs(currentCloseRoe) > (cutLossMultiplier*investValue)) || (cutLossMultiplier > 0.0 && currentHgLwRoe < 0 && math.Abs(currentHgLwRoe) > (cutLossMultiplier*investValue)) {
				// 	// log.Printf(fmt.Sprintf("datetime: %s|currentCloseRoe: %.1f|currentHgLwRoe: %.1f|investValue: %.1f", data.OpenTime[i], currentCloseRoe, currentHgLwRoe, investValue))
				// 	exitPrice := CalLastPriceFrmROE(currentEntry, investValue, int(margin), currentAction == "BUY", -100.0*cutLossMultiplier)
				// 	closeReason := "MANUAL"
				// 	pnl := calPNL(margin, currentEntry, exitPrice, investValue)
				// 	if currentAction == "SELL" {
				// 		pnl = calPNL(margin, exitPrice, currentEntry, investValue)
				// 	}
				// 	totalPNL += pnl
				// 	wallet += pnl
				// 	totalTrades++
				// 	if pnl > 0 {
				// 		wins++
				// 		results.WinLose[i] = 1
				// 	} else {
				// 		losses++
				// 		results.WinLose[i] = -1
				// 	}

				// 	results.Action[i] = currentAction
				// 	results.Entry[i] = currentEntry
				// 	results.Exit[i] = exitPrice
				// 	results.TakeProfit[i] = currentTakeProfit
				// 	results.StopLoss[i] = data.StopLoss[i]
				// 	results.PNL[i] = pnl
				// 	results.Balance[i] = wallet
				// 	results.ClosePosReason[i] = closeReason
				// 	inBuySell = false
				// } else if takeProfitCondition || stopLossCondition {
				// 	exitPrice := currentTakeProfit
				// 	closeReason := "hit TP"
				// 	if stopLossCondition {
				// 		exitPrice = data.StopLoss[i]
				// 		closeReason = "hit SL"
				// 	}
				// 	pnl := calPNL(margin, currentEntry, exitPrice, investValue)
				// 	if currentAction == "SELL" {
				// 		pnl = calPNL(margin, exitPrice, currentEntry, investValue)
				// 	}
				// 	totalPNL += pnl
				// 	wallet += pnl
				// 	totalTrades++
				// 	if pnl > 0 {
				// 		wins++
				// 		results.WinLose[i] = 1
				// 	} else {
				// 		losses++
				// 		results.WinLose[i] = -1
				// 	}

				// 	results.Action[i] = currentAction
				// 	results.Entry[i] = currentEntry
				// 	results.Exit[i] = exitPrice
				// 	results.TakeProfit[i] = currentTakeProfit
				// 	results.StopLoss[i] = data.StopLoss[i]
				// 	results.PNL[i] = pnl
				// 	results.Balance[i] = wallet
				// 	results.ClosePosReason[i] = closeReason
				// 	inBuySell = false
			} else {
				results.Action[i] = currentAction
				results.Entry[i] = currentEntry
				results.TakeProfit[i] = currentTakeProfit
				results.StopLoss[i] = data.StopLoss[i]
				// results.Balance[i] = wallet

			}
		} else if data.BUY[i] != 0 || data.SELL[i] != 0 {
			if data.BUY[i] != 0 {
				currentAction = "BUY"
			} else {
				currentAction = "SELL"
			}
			currentEntry = data.Close[i]
			currentTakeProfit = CalLastPriceFrmROE(currentEntry, investValue, int(margin), currentAction == "BUY", tpPercent)
			// currentStopLoss = data.StopLoss[i]
			inBuySell = true
			// results.Action[i] = currentAction
			// results.Entry[i] = currentEntry
		}

		results.Balance[i] = wallet
	}

	// common.SaveCsv4(results.ToMapSlice(), "detail")

	winRate := 0.0
	if totalTrades > 0 {
		winRate = float64(wins) / float64(totalTrades) * 100
	}

	// Create summary metrics struct
	summary := common.SummaryMetrics{
		Wins:         wins,
		Losses:       losses,
		TotalTrades:  totalTrades,
		WinRate:      winRate,
		TotalPNL:     totalPNL,
		FinalBalance: wallet,
	}

	return results, summary
}

func Backtestv201(data algorithm.SupertrendData2, margin float64, tpPercent float64, investAmt int, investPercent float64, cutLossMultiplier float64) (common.TradeResult, common.SummaryMetrics) {
	results := common.TradeResult{
		OpenTime:       data.OpenTime,
		Action:         make([]string, len(data.OpenTime)),
		Open:           data.Open,
		High:           data.High,
		Low:            data.Low,
		Close:          data.Close,
		HaOpen:         data.HaOpen,
		HaHigh:         data.HaHigh,
		HaLow:          data.HaLow,
		HaClose:        data.HaClose,
		Entry:          make([]float64, len(data.OpenTime)),
		Exit:           make([]float64, len(data.OpenTime)),
		TakeProfit:     make([]float64, len(data.OpenTime)),
		StopLoss:       data.StopLoss,
		WinLose:        make([]float64, len(data.OpenTime)),
		PNL:            make([]float64, len(data.OpenTime)),
		Balance:        make([]float64, len(data.OpenTime)),
		ClosePosReason: make([]string, len(data.OpenTime)),
	}
	var totalPNL, wallet float64
	var wins, losses, totalTrades int

	var inBuySell bool
	var currentAction string
	var currentEntry, currentTakeProfit float64

	wallet = float64(investAmt)
	investValue := float64(investAmt) * investPercent

	for i := 30; i < len(data.OpenTime); i++ {
		if data.Close[i] <= 0 || wallet <= 0 {
			continue
		}

		if inBuySell {
			takeProfitCondition := (currentAction == "BUY" && (data.Close[i] >= currentTakeProfit || data.High[i] >= currentTakeProfit)) ||
				(currentAction == "SELL" && (data.Close[i] <= currentTakeProfit || data.Low[i] <= currentTakeProfit))

			stopLossCondition := (currentAction == "BUY" && (data.Close[i] <= data.StopLoss[i] || data.Low[i] <= data.StopLoss[i])) ||
				(currentAction == "SELL" && (data.Close[i] >= data.StopLoss[i] || data.High[i] >= data.StopLoss[i]))

			currentCloseRoe := calPNL(margin, currentEntry, data.Close[i], investValue)
			currentHgLwRoe := calPNL(margin, currentEntry, data.Low[i], investValue)
			if currentAction == "SELL" {
				currentCloseRoe = calPNL(margin, data.Close[i], currentEntry, investValue)
				currentHgLwRoe = calPNL(margin, data.High[i], currentEntry, investValue)
			}

			if takeProfitCondition || stopLossCondition {
				exitPrice := currentTakeProfit
				closeReason := "hit TP"
				if stopLossCondition {
					exitPrice = data.StopLoss[i]
					closeReason = "hit SL"
				}
				pnl := calPNL(margin, currentEntry, exitPrice, investValue)
				if currentAction == "SELL" {
					pnl = calPNL(margin, exitPrice, currentEntry, investValue)
				}
				totalPNL += pnl
				wallet += pnl
				totalTrades++
				if pnl > 0 {
					wins++
					results.WinLose[i] = 1
				} else {
					losses++
					results.WinLose[i] = -1
				}

				results.Action[i] = currentAction
				results.Entry[i] = currentEntry
				results.Exit[i] = exitPrice
				results.TakeProfit[i] = currentTakeProfit
				results.StopLoss[i] = data.StopLoss[i]
				results.PNL[i] = pnl
				results.Balance[i] = wallet
				results.ClosePosReason[i] = closeReason
				inBuySell = false

				if data.SELL[i] > 0 {
					currentAction = "SELL"
					currentEntry = data.Close[i]
					currentTakeProfit = CalLastPriceFrmROE(currentEntry, investValue, int(margin), currentAction == "BUY", tpPercent)
					inBuySell = true
				}
				if data.BUY[i] > 0 {
					currentAction = "BUY"
					currentEntry = data.Close[i]
					currentTakeProfit = CalLastPriceFrmROE(currentEntry, investValue, int(margin), currentAction == "BUY", tpPercent)
					inBuySell = true
				}
			} else if (cutLossMultiplier > 0.0 && currentCloseRoe < 0 && math.Abs(currentCloseRoe) > (cutLossMultiplier*investValue)) || (cutLossMultiplier > 0.0 && currentHgLwRoe < 0 && math.Abs(currentHgLwRoe) > (cutLossMultiplier*investValue)) {
				// log.Printf(fmt.Sprintf("datetime: %s|currentCloseRoe: %.1f|currentHgLwRoe: %.1f|investValue: %.1f", data.OpenTime[i], currentCloseRoe, currentHgLwRoe, investValue))
				exitPrice := CalLastPriceFrmROE(currentEntry, investValue, int(margin), currentAction == "BUY", -100.0*cutLossMultiplier)
				closeReason := "MANUAL"
				pnl := calPNL(margin, currentEntry, exitPrice, investValue)
				if currentAction == "SELL" {
					pnl = calPNL(margin, exitPrice, currentEntry, investValue)
				}
				totalPNL += pnl
				wallet += pnl
				totalTrades++
				if pnl > 0 {
					wins++
					results.WinLose[i] = 1
				} else {
					losses++
					results.WinLose[i] = -1
				}

				results.Action[i] = currentAction
				results.Entry[i] = currentEntry
				results.Exit[i] = exitPrice
				results.TakeProfit[i] = currentTakeProfit
				results.StopLoss[i] = data.StopLoss[i]
				results.PNL[i] = pnl
				results.Balance[i] = wallet
				results.ClosePosReason[i] = closeReason
				inBuySell = false
			} else {
				results.Action[i] = currentAction
				results.Entry[i] = currentEntry
				results.TakeProfit[i] = currentTakeProfit
				results.StopLoss[i] = data.StopLoss[i]
			}
		} else if data.BUY[i] > 0 || data.SELL[i] > 0 {
			if data.BUY[i] != 0 {
				currentAction = "BUY"
			} else {
				currentAction = "SELL"
			}
			currentEntry = data.Close[i]
			currentTakeProfit = CalLastPriceFrmROE(currentEntry, investValue, int(margin), currentAction == "BUY", tpPercent)
			inBuySell = true
		}

		results.Balance[i] = wallet
	}

	// common.SaveCsv4(results.ToMapSlice(), "detail")

	winRate := 0.0
	if totalTrades > 0 {
		winRate = float64(wins) / float64(totalTrades) * 100
	}

	// Create summary metrics struct
	summary := common.SummaryMetrics{
		Wins:         wins,
		Losses:       losses,
		TotalTrades:  totalTrades,
		WinRate:      winRate,
		TotalPNL:     totalPNL,
		FinalBalance: wallet,
	}

	return results, summary
}

// BACKTEST

// func calPNL(margin, entryPrice, exitPrice, usdtValue float64) float64 {
// 	if usdtValue < 5 {
// 		usdtValue = 5
// 	}
// 	if entryPrice == 0 {
// 		return 0.0
// 	}
// 	buyAmt := usdtValue * margin
// 	quantity := buyAmt / entryPrice
// 	pnl := (exitPrice - entryPrice) * quantity

// 	return pnl
// }

// func calLastPriceFrmROE(entry float64, leverage int, inBuy bool, roe float64) float64 {
// 	positionMultiplier := 1.0
// 	if !inBuy {
// 		positionMultiplier = -1.0
// 	}
// 	return math.Round(((entry * roe) / (100 * float64(leverage)) * positionMultiplier) + entry)
// }
