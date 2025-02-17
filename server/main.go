package main

import (
	"context"
	"errors"
	"flag"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"sync/atomic"
	"syscall"
	"time"

	"pow-anti-ddos/app/common"
	"pow-anti-ddos/app/logx"
	"pow-anti-ddos/app/powx"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	address          = ":8080"
	maxHeaderBytes   = 1 << 10 // 1KB
	maxHeaderTimeout = 100 * time.Millisecond
)

var (
	difficulty        = flag.Int("difficulty", 10, "number of zero bits from 1 to 64")
	memoryCost        = flag.Uint("memoryCost", 1<<10, "memory allocation is required to prevent GPU attacks (default 1MB)")
	timeCost          = flag.Uint("timeCost", 1<<3, "cpu consumption level (default 8)")
	maxActiveRequests = flag.Int64("maxActiveRequests", 10, "we enable pow protection only on ddos attacks to make clients faster and less server side resource consumption")
	challengeTTL      = flag.Duration("challengeTTL", time.Minute, "challenge expiration is required to prevent postponed attacks")
)

var activeRequests int64

var wisdomQuotes = []string{
	"The only true wisdom is in knowing you know nothing. – Socrates",
	"Do what you can, with what you have, where you are. – Theodore Roosevelt",
	"Happiness depends upon ourselves. – Aristotle",
	"It is not length of life, but depth of life. – Ralph Waldo Emerson",
	"Knowing yourself is the beginning of all wisdom. – Aristotle",
	"You cannot teach a man anything; you can only help him find it within himself. – Galileo Galilei",
	"Success is not final, failure is not fatal: it is the courage to continue that counts. – Winston Churchill",
	"In three words I can sum up everything I've learned about life: it goes on. – Robert Frost",
	"We are what we repeatedly do. Excellence, then, is not an act, but a habit. – Will Durant (often attributed to Aristotle)",
	"All that we see or seem is but a dream within a dream. – Edgar Allan Poe",
}

func get(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
			return
		}
		logx.Info(r.Context(), "GET",
			zap.String("method", r.Method),
			zap.String("url", r.URL.Path),
			zap.String("ua", r.UserAgent()),
		)
		next.ServeHTTP(w, r)
	}
}

func pow(secretKey []byte, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get(common.HeaderXRequestID)
		if requestID == "" {
			requestID = uuid.NewString()
		}
		r = r.WithContext(context.WithValue(r.Context(), logx.ContextKeyRequestID, requestID))

		currentRequests := atomic.AddInt64(&activeRequests, 1)
		defer atomic.AddInt64(&activeRequests, -1)

		xff := r.Header.Get(common.HeaderXFF)
		if currentRequests > *maxActiveRequests && !isValidPoW(r, secretKey, xff) {
			challenge, err := powx.GenerateChallenge(requestID, secretKey, xff, powx.Params{
				Difficulty: *difficulty,
				MemoryCost: uint32(*memoryCost),
				TimeCost:   uint32(*timeCost),
				Threads:    uint8(runtime.NumCPU()),
			}, *challengeTTL)
			if err != nil {
				http.Error(w, "Failed to generate Challenge", http.StatusInternalServerError)
				return
			}
			w.Header().Set(common.HeaderPoWChallenge, challenge)
			http.Error(w, "Require Proof-of-Work solution", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	}
}

func isValidPoW(r *http.Request, secretKey []byte, xff string) bool {
	challenge := r.Header.Get(common.HeaderPoWChallenge)
	if challenge == "" {
		return false
	}

	params, err := powx.ParseAndVerifyChallenge(xff, challenge, secretKey)
	if err != nil {
		return false
	}

	solution := r.Header.Get(common.HeaderPoWSolution)
	if solution == "" {
		// NOTE: if challenge is not expired and xff is the same we can resend previous challenge
		return false
	}

	return powx.IsValidSolution(challenge, params, solution)
}

func quotes() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		quote := wisdomQuotes[rand.Intn(len(wisdomQuotes))]
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Wisdom: " + quote))
	}
}

func main() {
	defer logx.Sync()
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU())

	var secretKey = []byte(os.Getenv("SECRET_KEY")) // pass it over environment variables file, for security reasons

	mux := http.NewServeMux()

	mux.HandleFunc(common.PathQuotes, pow(secretKey, get(quotes())))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	server := &http.Server{
		Addr:              address,
		Handler:           mux,
		MaxHeaderBytes:    maxHeaderBytes,
		ReadHeaderTimeout: maxHeaderTimeout,
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
	}

	go func() {
		logx.Info(ctx, "Server listening on", zap.String("address", address))

		// NOTE: on production depends on environment we need to use tls
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logx.Fatal(ctx, "Failed to start", zap.Error(err))
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	logx.Info(ctx, "Shutting down server...")
	if err := server.Shutdown(ctx); err != nil {
		logx.Fatal(ctx, "Server forced to shutdown: %v", zap.Error(err))
	}

	logx.Info(ctx, "Server gracefully stopped")
}
