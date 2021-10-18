package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/CMPE281-Project1-GabrielChen/file-management-api/aws_usages"
	"github.com/CMPE281-Project1-GabrielChen/file-management-api/uuid"
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

type UploadFileReturn struct {
	FileID    string `json:"FileID"`
	UploadURL string `json:"UploadURL"`
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

	var body UploadFileRequest
	err := json.Unmarshal([]byte(request.Body), &body)
	if err != nil {
		return Response{StatusCode: 500}, fmt.Errorf("failed to unmarshall body\n")
	}

	signedUrl, err := aws_usages.SignURL("https://d3kp1rtsk23gz0.cloudfront.net/")
	if err != nil {
		return Response{StatusCode: 500}, fmt.Errorf("failed to sign url\n")
	}

	uuidWithHyphen := uuid.New()
	fileID := strings.Replace(uuidWithHyphen.String(), "-", "", -1)

	item := aws_usages.FileTableItem{
		FileID:   fileID,
		UserID:   userId,
		FileName: body.FileName,
		Modified: time.Now().String(),
	}

	if err := aws_usages.PutDynamoDB("dev-files", item); err != nil {
		return Response{StatusCode: 500}, fmt.Errorf("uploadFile failed: %v\n", err)
	}

	resp := UploadFileReturn{
		FileID:    fileID,
		UploadURL: signedUrl,
	}

	js, err := json.Marshal(resp)
	if err != nil {
		return Response{StatusCode: 500}, fmt.Errorf("failed to marshal signedURL\n")
	}

	return Response{
		StatusCode:      200,
		IsBase64Encoded: false,
		Body:            string(js),
	}, nil

	// var body UploadFileRequest

	// fmt.Printf("body: %v\n", request.Body)

	// _, params, err := mime.ParseMediaType(request.Headers["Content Type"])
	//
	// r := multipart.NewReader(strings.NewReader(request.Body), params["boundary"])
	//
	// f, err := r.ReadForm(32 << 20)
	// if err != nil {
	// return Response{StatusCode: 500}, fmt.Errorf("failed to read form\n")
	// }

	// f.

	// err := json.Unmarshal([]byte(request.Body), &body)
	// if err != nil {
	// return Response{StatusCode: 500}, fmt.Errorf("failed to unmarshall body\n")
	// }

	// sess := session.Must(session.NewSessionWithOptions(session.Options{
	// 	Config: aws.Config{Region: aws.String("us-west-2")},
	// }))

	// uploader := s3manager.NewUploader(sess)

	// _, err := uploader.Upload(&s3manager.UploadInput{
	// 	Bucket: aws.String("dev-user-file-store"),
	// 	Key:    aws.String("testing-upload"),
	// 	Body:   bytes.NewReader([]byte(request.Body)),
	// })

	// if err != nil {
	// 	return Response{StatusCode: 500}, fmt.Errorf("failed to upload file to s3, error: %v\n", err)
	// }

	// return Response{StatusCode: 200}, nil

	// Generate the File struct

	// Put in DynamoDB

	// Put in S3

	// Return

	// tableName := "dev-files"
	// fileID := "testing-file"

	// svc := dynamodb.New(session.New(),
	// 	aws.NewConfig().WithRegion("us-west-2"))

	// result, err := svc.GetItem(&dynamodb.GetItemInput{
	// 	TableName: aws.String(tableName),
	// 	Key: map[string]*dynamodb.AttributeValue{
	// 		"UserID": {
	// 			S: aws.String("testing-user"),
	// 		},
	// 	},
	// })

	// if err != nil {
	// 	fmt.Printf("failed to retrieve item from dynamodb %v\n", err)
	// }

	// if result.Item == nil {
	// 	fmt.Printf("item not found\n")
	// }

	// file := File{}

	// err = dynamodbattribute.UnmarshalMap(result.Item, &file)
	// if err != nil {
	// 	fmt.Printf("failed to unmarshal")
	// }

	// fmt.Println("UserID Item")
	// fmt.Printf("fileID: %v UserID: %v\n", file.FileID, file.UserID)

	// result, err = svc.GetItem(&dynamodb.GetItemInput{
	// 	TableName: aws.String(tableName),
	// 	Key: map[string]*dynamodb.AttributeValue{
	// 		"FileID": {
	// 			S: aws.String(fileID),
	// 		},
	// 	},
	// })

	// if err != nil {
	// 	fmt.Printf("failed to retrieve item from dynamodb %v\n", err)
	// }

	// if result.Item == nil {
	// 	fmt.Printf("item not found\n")
	// }

	// file = File{}

	// err = dynamodbattribute.UnmarshalMap(result.Item, &file)
	// if err != nil {
	// 	fmt.Printf("failed to unmarshal")
	// }

	// fmt.Println("FileID Item")
	// fmt.Printf("fileID: %v UserID: %v\n", file.FileID, file.UserID)

	// // userID := "testing-user"

	// privateKeyARN := "arn:aws:secretsmanager:us-west-2:988203901673:secret:dev-file-management-private-key-eZTVru"
	// privateKeyString, err := aws_usages.RetrieveSecret(privateKeyARN)
	// if err != nil {
	// 	fmt.Printf("failed to retrieve private key %v\n", privateKeyARN)
	// }

	// publicIDARN := "arn:aws:secretsmanager:us-west-2:988203901673:secret:dev-file-management-public-id-tyo5xL"
	// publicIDString, err := aws_usages.RetrieveSecret(publicIDARN)
	// if err != nil {
	// 	fmt.Printf("failed to retrieve publicID %v\n", publicIDARN)
	// }

	// rawURL := "https://d3kp1rtsk23gz0.cloudfront.net/test-image"
	// privateKey, err := sign.LoadPEMPrivKey(strings.NewReader(privateKeyString))
	// if err != nil {
	// 	fmt.Println("failed to load private key")
	// }

	// signer := sign.NewURLSigner(publicIDString, privateKey)
	// signedURL, err := signer.Sign(rawURL, time.Now().Add(1*time.Hour))
	// if err != nil {
	// 	log.Fatalf("Failed to sign url, err: %s\n", err.Error())
	// }

	// resp := Response{
	// 	StatusCode:      200,
	// 	IsBase64Encoded: false,
	// 	Body:            signedURL,
	// }

	// return resp, nil
}

func main() {
	lambda.Start(Handler)
}
