package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/prometheus/client_golang/prometheus/push"

	"github.com/namethatidesire/shrinkify-url/internal/metrics"
	"github.com/namethatidesire/shrinkify-url/internal/shortener"
	"github.com/namethatidesire/shrinkify-url/internal/store"
)

var (
	db			*store.DynamoStore
	sqsClient	*sqs.Client
	queueURL	string
	pushURL		string
)

func init() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalf("failed to load AWS config: %v", err)
	}

	db = store.New(dynamodb.NewFromConfig(cfg), os.Getenv("TABLE_NAME"))
	sqsClient = sqs.NewFromConfig(cfg)
	queueURL = os.Getenv("QUEUE_URL")
	pushURL = os.Getenv("GRAFANA_PUSH_URL")
}

func handler(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	reqCtxHTTP := req.RequestContext.HTTP
	switch {
	case reqCtxHTTP.Method == "POST" && reqCtxHTTP.Path == "/shorten":
		return handleShorten(ctx, req)
	case reqCtxHTTP.Method == "GET":
		code := strings.TrimPrefix(reqCtxHTTP.Path, "/")
		if code != "" {
			return handleRedirect(ctx, code)
		}
	}
	return response(http.StatusNotFound, "not found"), nil
}

func handleShorten(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	var body struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal([]byte(req.Body), &body); err != nil || body.URL == "" {
		return response(http.StatusBadRequest, "missing url"), nil
	}
	if err := shortener.ValidateURL(body.URL); err != nil {
		return response(http.StatusBadRequest, "invalid url"), nil
	}

	code, err := shortener.GenerateCode()
	if err != nil {
		return response(http.StatusInternalServerError, "could not generate code"), nil
	}
	if err := db.Put(ctx, code, body.URL); err != nil {
		return response(http.StatusInternalServerError, "could not store url"), nil
	}

	metrics.ShortenRequestsTotal.Inc()
	pushMetrics()

	return response(http.StatusOK, code), nil
}

func handleRedirect(ctx context.Context, code string) (events.APIGatewayV2HTTPResponse, error) {
	start := time.Now()

	record, err := db.Get(ctx, code)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return response(http.StatusNotFound, "not found"), nil
		}
		return response(http.StatusInternalServerError, "lookup failed"), nil
	}

	metrics.RedirectsTotal.WithLabelValues(code).Inc()
	metrics.RedirectLatency.Observe(float64(time.Since((start)).Seconds()))
	pushMetrics()

	// fire-and-forget SQS click event
	go sendClickEvent(code) // go --> goroutine; thread that runs concurrently

	return events.APIGatewayV2HTTPResponse{
		StatusCode:	http.StatusFound,
		Headers:	map[string]string{"Location": record.LongURL},
	}, nil
}

func sendClickEvent(code string) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()	// defer --> runs when the surrounding function returns, no matter how it exits;
					// cleans up context
	_, err := sqsClient.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:		&queueURL,
		MessageBody:	&code,
	})
	if err != nil {
		log.Printf("failed to send SQS message: %v", err)
		return
	}
	metrics.SQSMessagesSentTotal.Inc()
}

func pushMetrics() {
	if pushURL == "" {
		return
	}
	if err := push.New(pushURL, "shrinkify-url").Push(); err != nil {
		log.Printf("failed to push metrics: %v", err)
	}
}

func response(status int, body string) events.APIGatewayV2HTTPResponse {
	return events.APIGatewayV2HTTPResponse{
		StatusCode:	status,
		Headers:	map[string]string{"Content-Type": "application/json"},
		Body:		body,
	}
}

func main() {
	lambda.Start(handler)
}