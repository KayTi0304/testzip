package common

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"golang.org/x/time/rate"
)

type HostApiSecret struct {
	APIKey    string `json:"api_key"`
	SecretKey string `json:"secret_key"`
}

const HostApiSecretArnEnv = "HOST_API_SECRET"

var HostApiSecretArn = os.Getenv(HostApiSecretArnEnv)
var rateLimiter = rate.NewLimiter(rate.Every(1*time.Second), 4)

// RetrieveSecrets fetches the secret credentials from AWS Secrets Manager
func GetHostApiSecret() (HostApiSecret, error) {
	var secret HostApiSecret

	if HostApiSecretArn == "" {
		return secret, fmt.Errorf("error for host api - secret arn is required")
	}

	// Create a new AWS session
	sess := session.Must(session.NewSession())
	svc := secretsmanager.New(sess)

	// Retrieve the secret by its ARN
	result, err := svc.GetSecretValue(&secretsmanager.GetSecretValueInput{
		SecretId: aws.String(HostApiSecretArn),
	})
	if err != nil {
		return secret, fmt.Errorf("error for host api - failed to retrieve secret: %v", err)
	}

	// Unmarshal the secret string into Secret struct
	err = json.Unmarshal([]byte(*result.SecretString), &secret)
	if err != nil {
		return secret, fmt.Errorf("error for host api - failed to decode secret string: %v", err)
	}
	return secret, nil
}

func FetchBinanceFuturesKline(client *futures.Client, symbol, interval string, prevDaysFromNow int) ([]*futures.Kline, error) {
	// client := binance.NewFuturesClient(apiKey, secretKey)
	err := rateLimiter.Wait(context.Background())
	if err != nil {
		return nil, err
	}
	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Second) // Adjust the timeout as needed
	defer cancel()

	endTime := time.Now()
	startTime := endTime.AddDate(0, 0, -prevDaysFromNow)

	var allKlines []*futures.Kline
	for {
		klines, err := client.NewKlinesService().
			Symbol(symbol).
			Interval(interval).
			StartTime(startTime.UnixMilli()).
			EndTime(endTime.UnixMilli()).
			Limit(1000). // Max limit allowed by Binance
			Do(ctx)
		if err != nil {
			return allKlines, fmt.Errorf("error fetching futures price history: %v", err)
		}

		allKlines = append(allKlines, klines...)
		if len(klines) < 1000 {
			break // Break if the last request returned less than the limit, indicating it's the last batch
		}

		// Update startTime to the last kline's close time for the next batch
		lastKline := klines[len(klines)-1]
		startTime = time.UnixMilli(lastKline.CloseTime + 1)
	}

	return allKlines, nil
}

func FetchKlinesWithRetry(client *futures.Client, symbol string, interval string, prevDaysFromNow int) ([]*futures.Kline, error) {
	var klines []*futures.Kline
	var err error
	for retries := 0; retries < 5; retries++ {
		klines, err = FetchBinanceFuturesKline(client, symbol, interval, prevDaysFromNow)
		if err == nil {
			return klines, nil
		}
		log.Printf("Error fetching data for %s: %v. Retrying...", symbol, err)
		time.Sleep(time.Duration(retries*2) * time.Second) // Exponential backoff
	}
	return nil, fmt.Errorf("failed to fetch klines after retries: %w", err)
}

func GetActiveUSDTFuturesPairs(client *futures.Client) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second) // Adjust the timeout as needed
	defer cancel()
	exchangeInfo, err := client.NewExchangeInfoService().Do(ctx)
	fmt.Printf("err: %v\n", err)
	if err != nil {
		return nil, fmt.Errorf("error getting exchange info: %w", err)
	}

	var usdtPairs []string
	for _, symbol := range exchangeInfo.Symbols {
		if symbol.QuoteAsset == "USDT" && symbol.Status == "TRADING" {
			usdtPairs = append(usdtPairs, symbol.Symbol)
		}
	}

	return usdtPairs, nil
}
