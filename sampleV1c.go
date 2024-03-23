package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	v4 "github.com/aws/aws-sdk-go/aws/signer/v4"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/deoxxa/aws_signing_client"
	opensearchapi "github.com/opensearch-project/opensearch-go/v2/opensearchapi"
	"github.com/opensearch-project/opensearch-go/v3"
)

const IndexName = "go-test-index1"

func main() {
	if err := example(); err != nil {
		fmt.Println(fmt.Sprintf("Error: %s", err))
		os.Exit(1)
	}
}

const endpoint = "https://search-opensearch-fmi7qdkvmmpwtiqntnxnr32x7u.us-east-1.es.amazonaws.com" // e.g. https://opensearch-domain.region.com

func example() error {
	// ctx := context.Background()
	// Create an AWS request Signer and load AWS configuration using default config folder or env vars.
	// See https://docs.aws.amazon.com/opensearch-service/latest/developerguide/request-signing.html#request-signing-go

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String("us-east-1"),
		Credentials: credentials.NewSharedCredentials("", "abcd"),
	})
	_, err = sess.Config.Credentials.Get()
	fmt.Println(sess)
	fmt.Println(err)
	svc := iam.New(sess)
	result, err := svc.ListUsers(&iam.ListUsersInput{
		MaxItems: aws.Int64(10),
	})

	if err != nil {
		fmt.Println("Error", err)
		return err
	}

	for i, user := range result.Users {
		if user == nil {
			continue
		}
		fmt.Printf("%d user %s created %v\n", i, *user.UserName, user.CreateDate)
	}
	// return err

	credentialsEnv := credentials.NewEnvCredentials()
	signer1 := v4.NewSigner(credentialsEnv)
	fmt.Println(signer1)
	var DefaultTransport http.RoundTripper = &http.Transport{
		Dial: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 10 * time.Second,
	}
	client := &http.Client{Transport: DefaultTransport}

	signedClient, signerError := aws_signing_client.New(signer1, client, "es", "us-east-1")

	if signerError != nil {
		fmt.Errorf("Unable to create aws signed http transport client object: %v", signerError)
	}
	opensearchConfig := opensearch.Config{
		Addresses: []string{endpoint},
		//Username:  parameters.Username,
		//Password:  parameters.Password,
		//Username: "mercury",
		//Password: "",
		Transport: signedClient.Transport,
		//Transport: tp,
	}

	opensearchclient, err := opensearch.NewClient(opensearchConfig)

	fmt.Println(err)

	// Define index settings.
	settings := strings.NewReader(`{
		'settings': {
		  'index': {
			   'number_of_shards': 1,
			   'number_of_replicas': 2
			   }
			 }
		}`)

	// Create an index with non-default settings.
	res := opensearchapi.IndicesCreateRequest{
		Index: IndexName,
		Body:  settings,
	}
	fmt.Println("Creating index")
	fmt.Println(res)

	// Add a document to the index.
	document := strings.NewReader(`{
		   "title": "Moneyball",
		   "director": "Bennett Miller",
		   "year": "2011"
	   }`)

	docId := "1"
	req := opensearchapi.IndexRequest{
		Index:      IndexName,
		DocumentID: docId,
		Body:       document,
	}
	insertResponse, err := req.Do(context.Background(), opensearchclient)
	if err != nil {
		fmt.Println("failed to insert document ", err)
		os.Exit(1)
	}
	fmt.Println("Inserting a document")
	fmt.Println(insertResponse)
	defer insertResponse.Body.Close()

	// Search for the document.
	content := strings.NewReader(`{
		  "size": 5,
		  "query": {
			  "multi_match": {
			  "query": "miller",
			  "fields": ["title^2", "director"]
			  }
		 }
	   }`)

	search := opensearchapi.SearchRequest{
		Index: []string{IndexName},
		Body:  content,
	}

	searchResponse, err := search.Do(context.Background(), opensearchclient)
	if err != nil {
		fmt.Println("failed to search document ", err)
		os.Exit(1)
	}
	fmt.Println("Searching for a document")
	fmt.Println(searchResponse)
	defer searchResponse.Body.Close()

	// Delete the document.
	delete := opensearchapi.DeleteRequest{
		Index:      IndexName,
		DocumentID: docId,
	}

	deleteResponse, err := delete.Do(context.Background(), opensearchclient)
	if err != nil {
		fmt.Println("failed to delete document ", err)
		os.Exit(1)
	}
	fmt.Println("Deleting a document")
	fmt.Println(deleteResponse)
	defer deleteResponse.Body.Close()

	// Delete the previously created index.
	deleteIndex := opensearchapi.IndicesDeleteRequest{
		Index: []string{IndexName},
	}

	deleteIndexResponse, err := deleteIndex.Do(context.Background(), opensearchclient)
	if err != nil {
		fmt.Println("failed to delete index ", err)
		os.Exit(1)
	}
	fmt.Println("Deleting the index")
	fmt.Println(deleteIndexResponse)
	defer deleteIndexResponse.Body.Close()

	return nil
	// signer, err := requestsigner.NewSignerWithService(
	// 	session.Options{
	// 		// Specify profile to load for the session's config
	// 		// Profile: "abcd",

	// 		// Provide SDK Config options, such as Region.
	// 		Config: aws.Config{
	// 			Region: aws.String("us-east-1"),
	// 		},

	// 		// Force enable Shared Config support
	// 		SharedConfigState: session.SharedConfigEnable,
	// 	},
	// 	requestsigner.OpenSearchService, // Use requestsigner.OpenSearchServerless for Amazon OpenSearch Serverless.
	// )
	// if err != nil {
	// 	return err
	// }
	// Create an opensearch client and use the request-signer.
	// client, err := opensearchapi.NewClient(
	// 	opensearchapi.Config{
	// 		Client: opensearch.Config{
	// 			Addresses: []string{endpoint},
	// 			Signer:    signer,
	// 		},
	// 	},
	// )
	// if err != nil {
	// 	return err
	// }

	// ctx := context.Background()

	// ping, err := client.Ping(ctx, nil)
	// if err != nil {
	// 	return err
	// }

	// fmt.Println(ping)

	return nil
}
