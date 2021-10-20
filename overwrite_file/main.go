package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/CMPE281-Project1-GabrielChen/file-management-api/aws_usages"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

/****************Upload File Lambda********************/
// path: /{userId}
// steps 1:
// - put a new item in dynamoDB with a generated fields: UUID for fileID, string for TS
// step 2:
// - put item in s3 with the fileID
// return status 200 if all of these are accomplished, and return in body json with fields...

// Response is of type APIGatewayProxyResponse since we're leveraging the
// AWS Lambda Proxy Request functionality (default behavior)
//
// https://serverless.com/framework/docs/providers/aws/events/apigateway/#lambda-proxy-integration
type Response events.APIGatewayProxyResponse

type UploadFileRequest struct {
	FileName string `json:"FileName"`
}

type PatchFileReturn struct {
	PostURL string `json:"PostURL"`
}

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

	var body UploadFileRequest
	err := json.Unmarshal([]byte(request.Body), &body)
	if err != nil {
		return Response{StatusCode: 500}, fmt.Errorf("failed to unmarshall body\n")
	}

	signedUrl, err := aws_usages.SignURL("https://d3kp1rtsk23gz0.cloudfront.net/")
	if err != nil {
		return Response{StatusCode: 500}, fmt.Errorf("failed to sign url\n")
	}

	tableItem, err := aws_usages.GetFileDynamoDB("dev-files", fileID)
	if err != nil {
		return Response{StatusCode: 500}, err
	}
	if tableItem.UserID != userId {
		return Response{StatusCode: 404}, nil
	}

	t := time.Now().UTC().Format(time.RFC3339)

	item := aws_usages.OverwriteTableItem{
		FileName: body.FileName,
		Modified: t,
	}

	if err := aws_usages.OverwriteDynamoDB("dev-files", item, fileID); err != nil {
		return Response{StatusCode: 500}, fmt.Errorf("uploadFile failed: %v\n", err)
	}

	resp := PatchFileReturn{
		PostURL: signedUrl,
	}

	js, err := json.Marshal(resp)
	if err != nil {
		return Response{StatusCode: 500}, fmt.Errorf("failed to marshal signedURL\n")
	}

	return Response{
		StatusCode:      200,
		IsBase64Encoded: false,
		Headers: map[string]string{
			"Access-Control-Allow-Origin": "*",
		},
		Body: string(js),
	}, nil
}

func main() {
	lambda.Start(Handler)
}
