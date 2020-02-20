package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"

	"github.com/Azure/azure-sdk-for-go/services/resourcegraph/mgmt/2019-04-01/resourcegraph"
	"github.com/Azure/go-autorest/autorest/azure/auth"
)

type ResourceGraphQueryRequestInput struct {
	subscriptionID string
	query          string
	facets         []string
}

// Client is an API Client for Azure
type Client struct {
	SubscriptionID      string
	ResourceGraphClient resourcegraph.OperationsClient
}

// NewClient returns *Client with setting Authorizer
func NewClient(subscriptionID string) (*Client, error) {
	//a, err := auth.NewAuthorizerFromFile(azure.PublicCloud.ResourceManagerEndpoint)
	a, err := auth.NewAuthorizerFromCLI()
	if err != nil {
		return &Client{}, err
	}

	resourceGraphClient := resourcegraph.NewOperationsClient()
	resourceGraphClient.Authorizer = a

	return &Client{
		SubscriptionID:      subscriptionID,
		ResourceGraphClient: resourceGraphClient,
	}, nil
}

// FetchData is received by Resource Graph
type FetchData struct {
	// Column header
	Header []string
	// Data
	Data [][]string
}

func (data FetchData) OutputToFile(output *os.File) error {
	w := csv.NewWriter(output)
	if err := w.Write(data.Header); err != nil {
		return err
	}
	w.Flush()
	for _, elem := range data.Data {
		w.Write(elem)
		w.Flush()
	}
	return nil
}
func FetchResourceGraphData(c context.Context, client *Client, params ResourceGraphQueryRequestInput) (*FetchData, error) {
	var facetRequest []resourcegraph.FacetRequest
	for i := 0; i < len(params.facets); i++ {
		facetRequest = append(
			facetRequest,
			resourcegraph.FacetRequest{
				Expression: &params.facets[i],
			},
		)
	}
	request := &resourcegraph.QueryRequest{
		Subscriptions: &[]string{params.subscriptionID},
		Query:         &params.query,
		Options:       &resourcegraph.QueryRequestOptions{ResultFormat: resourcegraph.ResultFormatTable},
		Facets:        &facetRequest,
	}
	queryResponse, err := client.ResourceGraphClient.Resources(c, *request)
	if err != nil {
		return nil, err
	}

	columns := queryResponse.Data.(map[string]interface{})["columns"]
	rows := queryResponse.Data.(map[string]interface{})["rows"]

	header := []string{}
	for _, column := range columns.([]interface{}) {
		header = append(header, column.(map[string]interface{})["name"].(string))
	}
	results := [][]string{}
	// 取得したデータを1行ずつ処理
	for _, row := range rows.([]interface{}) {
		_result := []string{}
		// 1行を1カラムずつ処理
		for i, r := range row.([]interface{}) {
			// カラムの型に応じて処理を変える
			switch columns.([]interface{})[i].(map[string]interface{})["type"] {
			case "integer", "string":
				_result = append(_result, fmt.Sprint(r))
			case "object": // object の場合は JSON 化
				j, err := json.Marshal(r)
				if err != nil {
					return nil, err
				}
				_result = append(_result, string(j))
			}
		}
		// 1 行をカンマ区切りの1文字列とする
		results = append(results, _result)
	}
	return &FetchData{
		Header: header,
		Data:   results,
	}, nil
}
