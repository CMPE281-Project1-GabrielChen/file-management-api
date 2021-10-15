package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// Response is of type APIGatewayProxyResponse since we're leveraging the
// AWS Lambda Proxy Request functionality (default behavior)
//
// https://serverless.com/framework/docs/providers/aws/events/apigateway/#lambda-proxy-integration
type Response events.APIGatewayProxyResponse

// Handler is our lambda handler invoked by the `lambda.Start` function call
func Handler(ctx context.Context) (Response, error) {
	// hi := Response{
	// 	StatusCode:      200,
	// 	IsBase64Encoded: false,
	// 	Body:            "hi",
	// }

	// return hi, nil

	// instead of returning the file directly, return the URL for cloudfront?

	cloudfrontResp, err := http.Get("https://d3kp1rtsk23gz0.cloudfront.net/test-image")
	if err != nil {
		return Response{StatusCode: 404}, err
	}

	body, err := ioutil.ReadAll(cloudfrontResp.Body)
	if err != nil {
		return Response{
			StatusCode: 500,
			Body:       fmt.Sprintf("%v", err),
		}, err
	}

	resp := Response{
		StatusCode:      200,
		IsBase64Encoded: false,
		Body:            string(body),
		Headers: map[string]string{
			"Content-Disposition": "attachment; filename=testing-file-name",
			"Content-Type":        cloudfrontResp.Header.Get("Content-Type"),
		},
	}

	// body, err := json.Marshal(map[string]interface{}{
	// "message": "Go Serverless v1.0! Your function executed successfully!",
	// })
	// if err != nil {
	// return Response{StatusCode: 404}, err
	// }

	// json.HTMLEscape(&buf, body)

	// resp := Response{
	// StatusCode:      200,
	// IsBase64Encoded: false,
	// Body:            buf.String(),
	// Headers: map[string]string{
	// "Content-Disposition":
	// "X-MyCompany-Func-Reply": "hello-handler",
	// },
	// }

	return resp, nil
}

func main() {
	lambda.Start(Handler)
}
