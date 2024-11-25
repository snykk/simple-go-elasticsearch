package elastic

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/joho/godotenv"
)

var es *elasticsearch.Client

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("error loading .env file: %v", err)
	}

	caCertPath := os.Getenv("ELASTICSEARCH_CA_PATH")
	caCert, err := ioutil.ReadFile(caCertPath)
	if err != nil {
		log.Fatalf("error reading CA certificate: %v", err)
	}

	es, err = elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{os.Getenv("ELASTICSEARCH_URL")},
		Username:  os.Getenv("ELASTICSEARCH_USERNAME"),
		Password:  os.Getenv("ELASTICSEARCH_PASSWORD"),
		CACert:    caCert,
	})
	if err != nil {
		log.Fatalf("error creating Elasticsearch client: %v", err)
	}

	res, err := es.Info()
	if err != nil {
		log.Fatalf("error connecting to Elasticsearch: %v", err)
	}
	defer res.Body.Close()
	log.Println("Connected to Elasticsearch:", res)
}
