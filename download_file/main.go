package main

import (
	"context"
	"fmt"
	"net/url"

	"github.com/CMPE281-Project1-GabrielChen/file-management-api/aws_usages"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// https://serverless.com/framework/docs/providers/aws/events/apigateway/#lambda-proxy-integration
type Response events.APIGatewayProxyResponse

// Handler is our lambda handler invoked by the `lambda.Start` function call
func Handler(ctx context.Context, request events.APIGatewayProxyRequest) (Response, error) {
	// Get the file from the request, and get the UserID from the request
	userIdRaw, found := request.PathParameters["userId"]
	var userId string
	if found {
		value, err := url.QueryUnescape(userIdRaw)
		if nil != err {
			return Response{StatusCode: 500},
				fmt.Errorf("failed to unescape userIdRaw: %v, error: %v\n", userIdRaw, err)
		}

		userId = value
	} else {
		return Response{StatusCode: 500}, fmt.Errorf("no userId found???\n")
	}

	fileIdRaw, found := request.PathParameters["fileId"]
	var fileID string
	if found {
		value, err := url.QueryUnescape(fileIdRaw)
		if nil != err {
			return Response{StatusCode: 500},
				fmt.Errorf("failed to unescape fileIdRaw: %v, error: %v\n", fileIdRaw, err)
		}

		fileID = value
	} else {
		return Response{StatusCode: 500}, fmt.Errorf("no userId found???\n")
	}

	tableItem, err := aws_usages.GetFileDynamoDB("dev-files", fileID)
	if err != nil {
		return Response{StatusCode: 500}, err
	}

	if tableItem.UserID != userId {
		return Response{StatusCode: 404}, nil
	}

	signedUrl, err := aws_usages.SignURL(fmt.Sprintf("https://d3kp1rtsk23gz0.cloudfront.net/%s", fileID))
	if err != nil {
		return Response{StatusCode: 500}, fmt.Errorf("failed to sign url\n")
	}

	return Response{
		StatusCode:      307,
		IsBase64Encoded: false,
		Headers: map[string]string{
			"Location": signedUrl,
		},
	}, nil
}

func main() {
	lambda.Start(Handler)
}
