package store

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

var ErrNotFound = errors.New("code not found")

type URLRecord struct {
	Code	string	`dynamodbav:"code"`
	LongURL	string	`dynamodbav:"long_url"`
	Clicks	int		`dynamodbav:"clicks"`
}

type DynamoStore struct {
	client		*dynamodb.Client
	tableName	string
}

func New(client *dynamodb.Client, tableName string) *DynamoStore {
	return &DynamoStore{client: client, tableName: tableName}
}

func (s *DynamoStore) Put(ctx context.Context, code string, longURL string) error {
	record := URLRecord{Code: code, LongURL: longURL}
	item, err := attributevalue.MarshalMap(record)
	if err != nil {
		return err
	}
	_, err = s.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: 	aws.String(s.tableName),
		Item:		item,
	})
	return err
}

func (s *DynamoStore) Get(ctx context.Context, code string) (URLRecord, error) {
	out, err := s.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(s.tableName),
		Key: map[string]types.AttributeValue{
			"code": &types.AttributeValueMemberS{Value: code},
		},
	})
	if err != nil {
		return URLRecord{}, err
	}
	if out.Item == nil {
		return URLRecord{}, ErrNotFound
	}
	var record URLRecord
	err = attributevalue.UnmarshalMap(out.Item, &record)
	return record, err
}