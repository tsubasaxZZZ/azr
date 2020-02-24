package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"

	"github.com/Azure/azure-sdk-for-go/services/resourcegraph/mgmt/2019-04-01/resourcegraph"
	"github.com/Azure/azure-sdk-for-go/services/resourcegraph/mgmt/2019-04-01/resourcegraph/resourcegraphapi"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/azure/auth"
)

type ResourceGraphQueryRequestInput struct {
	subscriptionID string
	query          string
	facets         []string
	skipToken      string
	skip           int32
}

// Client is an API Client for Azure
type Client struct {
	SubscriptionID      string
	ResourceGraphClient resourcegraphapi.BaseClientAPI
}

// NewClient returns *Client with setting Authorizer
func NewClient(subscriptionID string) (*Client, error) {
	a, err := auth.NewAuthorizerFromCLI()
	if err != nil {
		a, err = auth.NewAuthorizerFromFile(azure.PublicCloud.ResourceManagerEndpoint)
		if err != nil {
			return &Client{}, err
		}
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
	var fetchData FetchData
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
		Options:       &resourcegraph.QueryRequestOptions{ResultFormat: resourcegraph.ResultFormatTable, SkipToken: &params.skipToken, Skip: &params.skip},
		Facets:        &facetRequest,
	}
	queryResponse, err := client.ResourceGraphClient.Resources(c, *request)
	if err != nil {
		return nil, err
	}

	if queryResponse.SkipToken != nil {
		params.skipToken = *queryResponse.SkipToken
		params.skip += int32(*queryResponse.Count)
		_data, err := FetchResourceGraphData(c, client, params)
		if err != nil {
			return nil, err
		}
		for _, elem := range _data.Data {
			fetchData.Data = append(fetchData.Data, elem)
		}
	}
	columns := queryResponse.Data.(map[string]interface{})["columns"]
	rows := queryResponse.Data.(map[string]interface{})["rows"]

	for _, column := range columns.([]interface{}) {
		fetchData.Header = append(fetchData.Header, column.(map[string]interface{})["name"].(string))
	}
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
		fetchData.Data = append(fetchData.Data, _result)
	}
	return &fetchData, nil
}
