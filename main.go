package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
)

type webhook struct {
	Action      string `json:"action"`
	PullRequest struct {
		CreatedAt    string  `json:"created_at"`
		ClosedAt     *string `json:"closed_at"`
		Additions    int     `json:"additions"`
		Deletions    int     `json:"deletions"`
		ChangedFiles int     `json:"changed_files"`
		Base         struct {
			Repo struct {
				Name string `json:"full_name"`
			} `json:"repo"`
		} `json:"base"`
	} `json:"pull_request"`
}

func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	rawToken := os.Getenv("WEBHOOK_SECRET_TOKEN")
	mac := hmac.New(sha1.New, []byte(rawToken))
	mac.Write([]byte(request.Body))
	expectedMAC := mac.Sum(nil)
	expectedHubSignature := fmt.Sprintf("sha1=%s", hex.EncodeToString(expectedMAC))
	//fmt.Println("Headers:")
	//for key, value := range request.Headers {
	//	fmt.Printf("    %s: %s\n", key, value)
	//}

	if request.Headers["X-Hub-Signature"] != expectedHubSignature {
		return events.APIGatewayProxyResponse{Body: "Invalid Signature", StatusCode: 400}, nil
	}

	if request.Headers["X-Github-Event"] != "pull_request" {
		// ignore test webhooks and events that are not PRs
		fmt.Printf("ignoring X-GitHub-Event %s\n", request.Headers["X-Github-Event"])
		return events.APIGatewayProxyResponse{Body: "LOVE, TIME, IDEAS", StatusCode: 200}, nil
	}

	hook := webhook{}
	err := json.Unmarshal([]byte(request.Body), &hook)
	if err != nil {
		fmt.Println(err)
		return events.APIGatewayProxyResponse{Body: "Parse error. Fast asleep and finished with the world.", StatusCode: 400}, nil
	}

	if hook.Action != "closed" || hook.PullRequest.ClosedAt == nil {
		// this PR isn't closed yet, we'll ignore it and only
		// record stats on PRs as they are closed.
		return events.APIGatewayProxyResponse{Body: "nothingness haunts being", StatusCode: 200}, nil
	}

	fmt.Println(hook)
	t1, e := time.Parse(
		time.RFC3339,
		hook.PullRequest.CreatedAt)
	if e != nil {
		fmt.Println(e)
	}

	t2, e := time.Parse(
		time.RFC3339,
		*hook.PullRequest.ClosedAt)
	if e != nil {
		fmt.Println(e)
	}

	duration := t2.Sub(t1).Round(time.Second)
	seconds := duration.Seconds()

	fmt.Printf("PR open for %d seconds\n", seconds)
	size := hook.PullRequest.Additions + hook.PullRequest.Deletions
	fmt.Printf("PR had %d lines added/removed\n", size)
	fmt.Printf("PR had %d changed files\n", hook.PullRequest.ChangedFiles)

	// post metric to Cloudwatch
	// regular 1 minute resolution metric
	// in the PRStats namespace
	// Unit is count, this is 1 PR
	// metrics are count, duration, size (additions +  deletions)
	repoDim := cloudwatch.Dimension{}
	repoDim.SetName("repo")
	repoDim.SetValue(hook.PullRequest.Base.Repo.Name)
	dimensions := []*cloudwatch.Dimension{}
	dimensions = append(dimensions, &repoDim)

	countDataPoint := cloudwatch.MetricDatum{}
	countDataPoint.SetMetricName("prcount")
	countDataPoint.SetUnit("Count")
	countDataPoint.SetValue(float64(1))
	countDataPoint.SetDimensions(dimensions)

	durationDataPoint := cloudwatch.MetricDatum{}
	durationDataPoint.SetMetricName("prduration")
	durationDataPoint.SetUnit("Seconds")
	durationDataPoint.SetValue(seconds)
	durationDataPoint.SetDimensions(dimensions)

	sizeDataPoint := cloudwatch.MetricDatum{}
	sizeDataPoint.SetMetricName("prsize")
	sizeDataPoint.SetUnit("Count")
	sizeDataPoint.SetValue(float64(size))
	sizeDataPoint.SetDimensions(dimensions)

	data := []*cloudwatch.MetricDatum{}
	data = append(data, &countDataPoint)
	data = append(data, &durationDataPoint)
	data = append(data, &sizeDataPoint)
	namespace := "SSDL"
	input := cloudwatch.PutMetricDataInput{MetricData: data, Namespace: &namespace}

	sess := session.Must(session.NewSession())
	svc := cloudwatch.New(sess)

	_, err = svc.PutMetricData(&input)
	if err != nil {
		fmt.Println(err)
	}

	return events.APIGatewayProxyResponse{Body: "The heart of metaphor is inference.", StatusCode: 200}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
