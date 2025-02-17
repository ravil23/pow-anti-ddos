package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"sync"
	"time"

	"pow-anti-ddos/app/common"
	"pow-anti-ddos/app/logx"
	"pow-anti-ddos/app/powx"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

var (
	clientsCount  = flag.Int("concurrency", 100, "concurrent clients count")
	requestsCount = flag.Int("requests", 10, "per client requests count")
	serverURL     = flag.String("serverURL", "http://localhost:8080", "server url")
)

func solveChallenge(ctx context.Context, challenge string) (string, error) {
	params, err := powx.ParseUnverifiedChallenge(challenge)
	if err != nil {
		return "", err
	}

	candidateInt := 0
	start := time.Now()
	for {
		candidateStr := strconv.Itoa(candidateInt)
		if powx.IsValidSolution(challenge, params, candidateStr) {
			totalDuration := time.Since(start)
			logx.Info(ctx, "Solved challenge", zap.Int("candidate", candidateInt), zap.Duration("total_duration", totalDuration), zap.Float64("avg_duration", totalDuration.Seconds()/float64(candidateInt+1)))
			return candidateStr, nil
		}
		candidateInt++
	}
}

func getQuotes(_ context.Context, client *http.Client, url string, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/%s", url, common.PathQuotes), nil)
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		req.Header.Add(k, v)
	}
	return client.Do(req)
}

func callServer(ctx context.Context, client *http.Client, requestID string) error {
	resp, err := getQuotes(ctx, client, *serverURL, map[string]string{
		common.HeaderXRequestID: requestID,
	})
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusForbidden {
		challenge := resp.Header.Get(common.HeaderPoWChallenge)
		solution, err := solveChallenge(ctx, challenge)
		if err != nil {
			return err
		}
		resp, err = getQuotes(ctx, client, *serverURL, map[string]string{
			common.HeaderPoWChallenge: challenge,
			common.HeaderPoWSolution:  solution,
			common.HeaderXRequestID:   requestID,
		})
		if err != nil {
			return err
		}
	}
	reader := bufio.NewReader(resp.Body)
	line, _, err := reader.ReadLine()
	if err != nil {
		return err
	}

	logx.Info(ctx, "Server response", zap.ByteString("line", line))
	return nil
}

func main() {
	defer logx.Sync()
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU())

	ctx := context.Background()
	logx.Info(ctx, "Starting client...")
	client := http.DefaultClient

	wg := sync.WaitGroup{}
	for i := 0; i < *clientsCount; i++ {
		wg.Add(1)
		clientCtx := context.WithValue(ctx, logx.ContextKeyClientNum, fmt.Sprintf("%d", i))
		go func() {
			defer wg.Done()
			logx.Info(clientCtx, "Starting client")
			for j := 0; j < *requestsCount; j++ {
				requestID := uuid.NewString()
				clientCtx = context.WithValue(clientCtx, logx.ContextKeyRequestID, requestID)
				if err := callServer(clientCtx, client, requestID); err != nil {
					logx.Error(clientCtx, "Failed to call server", zap.Error(err))
				}
			}
			logx.Info(clientCtx, "Finishing client")
		}()
	}
	wg.Wait()
}
