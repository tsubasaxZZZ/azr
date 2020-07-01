package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
	"gopkg.in/yaml.v2"
)

const (
	// QueryConcurrency is number of query concurrency
	QueryConcurrency = 5
)

func main() {
	app := &cli.App{
		Name:   "azr",
		Usage:  "Azure Resource Graph Command",
		Action: GetResources,
		Before: before,
	}
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:     "subscriptionID",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "query",
			Aliases:  []string{"q"},
			Required: true,
		},
		&cli.StringFlag{
			Name:    "file",
			Aliases: []string{"f"},
			Usage:   "Speify output filepath(If not specify, out to stdout)",
		},
		//&cli.BoolFlag{
		//Name:    "verbose",
		//Aliases: []string{"v"},
		//Usage:   "Output verbose logs to stderr",
		//},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
func before(c *cli.Context) error {
	//log.SetOutput(os.Stderr)
	return nil
}

type QueryConfig struct {
	Name   string `yaml:"name"`
	Query  string `yaml:"query"`
	Output *os.File
}

func validationYAMLConfig(config *[]QueryConfig) error {
	// name の重複チェック
	names := map[string]int{}
	for i := 0; i < len(*config); i++ {
		names[(*config)[i].Name]++
	}
	for k, v := range names {
		if v > 1 {
			return fmt.Errorf("%s is duplicate. Please change other name", k)
		}
	}
	return nil
}
func GetResources(c *cli.Context) error {
	client, err := NewClient(c.String("subscriptionID"))
	if err != nil {
		return err
	}
	q := c.String("query")

	var qcs []QueryConfig
	// クエリ文字列の先頭文字が @ の場合はファイル名としてクエリを読み込む
	if strings.Index(q, "@") == 0 {
		queryFilePath := q[1:]
		log.Printf("Load config from %s\n", queryFilePath)
		configData, err := ioutil.ReadFile(queryFilePath)
		if err != nil {
			return err
		}
		if err := yaml.UnmarshalStrict(configData, &qcs); err != nil {
			return err
		}

		// YAML のバリデーション
		if err := validationYAMLConfig(&qcs); err != nil {
			return err
		}

		// <name>.csv で結果を保存する
		for i := 0; i < len(qcs); i++ {
			qcs[i].Output, err = os.Create(qcs[i].Name + ".csv")
			if err != nil {
				return err
			}
			defer qcs[i].Output.Close()
		}
	} else {
		var qc QueryConfig
		qc.Query = q
		qc.Output = os.Stdout
		f := c.String("file")
		if f != "" {
			filePath, err := os.Create(f)
			if err != nil {
				return err
			}
			qc.Output = filePath
		}
		qcs = append(qcs, qc)

	}

	eg := errgroup.Group{}
	s := semaphore.NewWeighted(QueryConcurrency)
	for _, qc := range qcs {
		s.Acquire(context.Background(), 1)

		qc := qc

		eg.Go(
			func() error {
				defer s.Release(1)

				regexNewLine := regexp.MustCompile(`\r\n|\r|\n`)
				log.Printf("Get resource graph:Name=[%s],Query=[%s]", qc.Name, regexNewLine.ReplaceAllString(qc.Query, ""))
				data, errGet := getResourceGraphData(c, &qc, client)
				if errGet != nil {
					return fmt.Errorf("Name: %s, Error: %s", qc.Name, errGet)
				}
				if err := data.OutputToFile(qc.Output); err != nil {
					return err
				}
				return nil
			})

	}

	if err := eg.Wait(); err != nil {
		log.Fatal(err)
	}

	return nil
}

func getResourceGraphData(c *cli.Context, qc *QueryConfig, client *Client) (*FetchData, error) {
	var data *FetchData
	qr := &ResourceGraphQueryRequestInput{
		subscriptionID: c.String("subscriptionID"),
		query:          qc.Query,
	}
	data, err := FetchResourceGraphData(context.TODO(), client, *qr)
	if err != nil {
		return nil, err
	}
	return data, nil
}
