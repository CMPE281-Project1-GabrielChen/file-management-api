package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudfront/sign"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

// Response is of type APIGatewayProxyResponse since we're leveraging the
// AWS Lambda Proxy Request functionality (default behavior)
//
// https://serverless.com/framework/docs/providers/aws/events/apigateway/#lambda-proxy-integration
type Response events.APIGatewayProxyResponse

type File struct {
	UserID   string
	FileID   string
	FileName string
	Modified string
	Location string
}

// Handler is our lambda handler invoked by the `lambda.Start` function call
func Handler(ctx context.Context) (Response, error) {
	tableName := "dev-files"
	fileID := "testing-file"

	svc := dynamodb.New(session.New(),
		aws.NewConfig().WithRegion("us-west-2"))

	result, err := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"UserID": {
				S: aws.String("testing-user"),
			},
		},
	})

	if err != nil {
		fmt.Printf("failed to retrieve item from dynamodb %v\n", err)
	}

	if result.Item == nil {
		fmt.Printf("item not found\n")
	}

	file := File{}

	err = dynamodbattribute.UnmarshalMap(result.Item, &file)
	if err != nil {
		fmt.Printf("failed to unmarshal")
	}

	fmt.Println("UserID Item")
	fmt.Printf("fileID: %v UserID: %v\n", file.FileID, file.UserID)

	result, err = svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"FileID": {
				S: aws.String(fileID),
			},
		},
	})

	if err != nil {
		fmt.Printf("failed to retrieve item from dynamodb %v\n", err)
	}

	if result.Item == nil {
		fmt.Printf("item not found\n")
	}

	file = File{}

	err = dynamodbattribute.UnmarshalMap(result.Item, &file)
	if err != nil {
		fmt.Printf("failed to unmarshal")
	}

	fmt.Println("FileID Item")
	fmt.Printf("fileID: %v UserID: %v\n", file.FileID, file.UserID)

	// userID := "testing-user"

	privateKeyARN := "arn:aws:secretsmanager:us-west-2:988203901673:secret:dev-file-management-private-key-eZTVru"
	privateKeyString, err := RetrieveSecret(privateKeyARN)
	if err != nil {
		fmt.Printf("failed to retrieve private key %v\n", privateKeyARN)
	}

	publicIDARN := "arn:aws:secretsmanager:us-west-2:988203901673:secret:dev-file-management-public-id-tyo5xL"
	publicIDString, err := RetrieveSecret(publicIDARN)
	if err != nil {
		fmt.Printf("failed to retrieve publicID %v\n", publicIDARN)
	}

	rawURL := "https://d3kp1rtsk23gz0.cloudfront.net/"
	privateKey, err := sign.LoadPEMPrivKey(strings.NewReader(privateKeyString))
	if err != nil {
		fmt.Println("failed to load private key")
	}

	signer := sign.NewURLSigner(publicIDString, privateKey)
	signedURL, err := signer.Sign(rawURL, time.Now().Add(1*time.Hour))
	if err != nil {
		log.Fatalf("Failed to sign url, err: %s\n", err.Error())
	}

	resp := Response{
		StatusCode:      200,
		IsBase64Encoded: false,
		Body:            signedURL,
	}

	return resp, nil
}

func main() {
	lambda.Start(Handler)
}

func RetrieveSecret(secretName string) (string, error) {
	region := "us-west-2"

	//Create a Secrets Manager client
	svc := secretsmanager.New(session.New(),
		aws.NewConfig().WithRegion(region))
	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(secretName),
		VersionStage: aws.String("AWSCURRENT"), // VersionStage defaults to AWSCURRENT if unspecified
	}

	// In this sample we only handle the specific exceptions for the 'GetSecretValue' API.
	// See https://docs.aws.amazon.com/secretsmanager/latest/apireference/API_GetSecretValue.html

	result, err := svc.GetSecretValue(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case secretsmanager.ErrCodeDecryptionFailure:
				// Secrets Manager can't decrypt the protected secret text using the provided KMS key.
				fmt.Println(secretsmanager.ErrCodeDecryptionFailure, aerr.Error())

			case secretsmanager.ErrCodeInternalServiceError:
				// An error occurred on the server side.
				fmt.Println(secretsmanager.ErrCodeInternalServiceError, aerr.Error())

			case secretsmanager.ErrCodeInvalidParameterException:
				// You provided an invalid value for a parameter.
				fmt.Println(secretsmanager.ErrCodeInvalidParameterException, aerr.Error())

			case secretsmanager.ErrCodeInvalidRequestException:
				// You provided a parameter value that is not valid for the current state of the resource.
				fmt.Println(secretsmanager.ErrCodeInvalidRequestException, aerr.Error())

			case secretsmanager.ErrCodeResourceNotFoundException:
				// We can't find the resource that you asked for.
				fmt.Println(secretsmanager.ErrCodeResourceNotFoundException, aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return "", err
	}

	// Decrypts secret using the associated KMS CMK.
	// Depending on whether the secret is a string or binary, one of these fields will be populated.
	var secretString string
	// decodedBinarySecret string
	if result.SecretString != nil {
		secretString = *result.SecretString
	} else {
		return "", fmt.Errorf("invalid secret")
	}

	return secretString, nil
}
