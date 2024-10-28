package common

type TradeResult struct {
	OpenTime       []string
	Action         []string
	Open           []float64
	High           []float64
	Low            []float64
	Close          []float64
	HaOpen         []float64
	HaHigh         []float64
	HaLow          []float64
	HaClose        []float64
	Entry          []float64
	Exit           []float64
	TakeProfit     []float64
	StopLoss       []float64
	WinLose        []float64
	PNL            []float64
	Balance        []float64
	ClosePosReason []string
}

type SummaryMetrics struct {
	Wins         int
	Losses       int
	TotalTrades  int
	WinRate      float64
	TotalPNL     float64
	FinalBalance float64
}

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func (data *TradeResult) ToMapSlice() []map[string]interface{} {
	length := len(data.OpenTime) // Assuming all slices are the same length
	results := make([]map[string]interface{}, length)

	for i := 0; i < length; i++ {
		row := map[string]interface{}{
			"OpenTime":       data.OpenTime[i],
			"Action":         data.Action[i],
			"Open":           data.Open[i],
			"High":           data.High[i],
			"Low":            data.Low[i],
			"Close":          data.Close[i],
			"Entry":          data.Entry[i],
			"Exit":           data.Exit[i],
			"TakeProfit":     data.TakeProfit[i],
			"StopLoss":       data.StopLoss[i],
			"WinLose":        data.WinLose[i],
			"PNL":            data.PNL[i],
			"Balance":        data.Balance[i],
			"ClosePosReason": data.ClosePosReason[i],
		}
		results[i] = row
	}

	return results
}
