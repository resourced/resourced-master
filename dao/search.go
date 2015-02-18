package dao

import (
	"github.com/awslabs/aws-sdk-go/aws"
	"github.com/awslabs/aws-sdk-go/gen/cloudsearch"
)

func NewSearch(accessKey, secretKey, securityToken, region string) *Search {
	s := &Search{}
	s.credentials = aws.Creds(accessKey, secretKey, securityToken)
	s.cloudsearch = cloudsearch.New(s.credentials, region, nil)

	return s
}

type Search struct {
	credentials aws.CredentialsProvider
	cloudsearch *cloudsearch.CloudSearch
}
