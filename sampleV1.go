package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws/session"
	requestsigner "github.com/opensearch-project/opensearch-go/v3/signer/aws"

	"github.com/opensearch-project/opensearch-go/v3"
	"github.com/opensearch-project/opensearch-go/v3/opensearchapi"
)

const IndexName = "go-test-index1"

func main() {
	if err := example(); err != nil {
		fmt.Println(fmt.Sprintf("Error: %s", err))
		os.Exit(1)
	}
}

const endpoint = "url" // e.g. https://opensearch-domain.region.com

func example() error {
	// Create an AWS request Signer and load AWS configuration using default config folder or env vars.
	// See https://docs.aws.amazon.com/opensearch-service/latest/developerguide/request-signing.html#request-signing-go
	signer, err := requestsigner.NewSignerWithService(
		session.Options{Profile: "default"},
		requestsigner.OpenSearchService, // Use requestsigner.OpenSearchServerless for Amazon OpenSearch Serverless.
	)
	if err != nil {
		return err
	}
	// Create an opensearch client and use the request-signer.
	client, err := opensearchapi.NewClient(
		opensearchapi.Config{
			Client: opensearch.Config{
				Addresses: []string{endpoint},
				Signer:    signer,
			},
		},
	)
	if err != nil {
		return err
	}

	ctx := context.Background()

	ping, err := client.Ping(ctx, nil)
	if err != nil {
		return err
	}

	fmt.Println(ping)

	return nil
}
