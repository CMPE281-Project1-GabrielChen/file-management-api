package aws_usages

import (
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudfront/sign"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

type FileTableItem struct {
	FileID   string `json:"FileID"`
	UserID   string `json:"UserID"`
	FileName string `json:"FileName"`
	Modified string `json:"Modified"`
}

func ListAllFilesDynamoDB(tableName string) (*[]FileTableItem, error) {
	svc := dynamodb.New(session.New(),
		aws.NewConfig().WithRegion("us-west-2"))

	params := &dynamodb.ScanInput{
		TableName: aws.String(tableName),
	}

	result, err := svc.Scan(params)
	if err != nil {
		return nil, fmt.Errorf("query api call failed: %s", err)
	}

	var files []FileTableItem

	for _, i := range result.Items {
		f := FileTableItem{}
		err = dynamodbattribute.UnmarshalMap(i, &f)
		if err != nil {
			return nil, fmt.Errorf("Got error unmarshalling: %s", err)
		}

		files = append(files, f)
	}

	return &files, nil
}

func ListFilesDynamoDB(tableName string, userID string) (*[]FileTableItem, error) {
	svc := dynamodb.New(session.New(),
		aws.NewConfig().WithRegion("us-west-2"))

	filt := expression.Name("UserID").Equal(expression.Value(userID))

	expr, err := expression.NewBuilder().WithFilter(filt).Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build expression: %s", err)
	}

	params := &dynamodb.ScanInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		TableName:                 aws.String(tableName),
	}

	result, err := svc.Scan(params)
	if err != nil {
		return nil, fmt.Errorf("query api call failed: %s", err)
	}

	var files []FileTableItem

	for _, i := range result.Items {
		f := FileTableItem{}
		err = dynamodbattribute.UnmarshalMap(i, &f)
		if err != nil {
			return nil, fmt.Errorf("Got error unmarshalling: %s", err)
		}

		files = append(files, f)
	}

	return &files, nil
}

func DeleteDynamoDB(tableName string, fileID string) error {
	svc := dynamodb.New(session.New(),
		aws.NewConfig().WithRegion("us-west-2"))

	_, err := svc.DeleteItem(&dynamodb.DeleteItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"FileID": {
				S: aws.String(fileID),
			},
		},
	})
	if err != nil {
		return fmt.Errorf("dynamodb responded with error: %v, error: %v\n", tableName, err)
	}

	return nil
}

func GetFileDynamoDB(tableName string, fileID string) (*FileTableItem, error) {
	svc := dynamodb.New(session.New(),
		aws.NewConfig().WithRegion("us-west-2"))

	result, err := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"FileID": {
				S: aws.String(fileID),
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query dynamodb tableName: %v, error: %v\n", tableName, err)
	}
	if result.Item == nil {
		return nil, fmt.Errorf("item not found tableName: %v, fileId: %v\n", fileID)
	}
	file := FileTableItem{}

	err = dynamodbattribute.UnmarshalMap(result.Item, &file)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal\n")
	}

	return &file, nil
}

func PutDynamoDB(tableName string, fileData FileTableItem) error {
	svc := dynamodb.New(session.New(),
		aws.NewConfig().WithRegion("us-west-2"))

	dynamoItem, err := dynamodbattribute.MarshalMap(fileData)
	if err != nil {
		return fmt.Errorf("failed to marshal fileData into dynamoItem %v\n", fileData)
	}

	input := &dynamodb.PutItemInput{
		Item:      dynamoItem,
		TableName: aws.String(tableName),
	}

	_, err = svc.PutItem(input)
	if err != nil {
		return fmt.Errorf("PutItem error: %v\n", err)
	}

	return nil
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

func SignURL(rawURL string) (string, error) {
	privateKeyARN := "arn:aws:secretsmanager:us-west-2:988203901673:secret:dev-file-management-private-key-eZTVru"
	privateKeyString, err := RetrieveSecret(privateKeyARN)
	if err != nil {
		fmt.Printf("failed to retrieve private key %v\n", privateKeyARN)
	}

	publicIDARN := "arn:aws:secretsmanager:us-west-2:988203901673:secret:dev-file-management-public-id-tyo5xL"
	publicIDString, err := RetrieveSecret(publicIDARN)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve publicID %v", publicIDARN)
	}

	privateKey, err := sign.LoadPEMPrivKey(strings.NewReader(privateKeyString))
	if err != nil {
		return "", fmt.Errorf("failed to parse primary key")
	}

	signer := sign.NewURLSigner(publicIDString, privateKey)
	signedURL, err := signer.Sign(rawURL, time.Now().Add(1*time.Hour))
	if err != nil {
		return "", fmt.Errorf("failed to sign url")
	}

	return signedURL, nil
}
