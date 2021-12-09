package kafka

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type JsonProperty map[string]string

type AvroSchema struct {
	Namespace  string                `json:"namespace"`
	Type       string                `json:"type"`
	SchemaName string                `json:"name"`
	Fields     [](map[string]string) `json:"fields"`
}

type JsonSchema struct {
	Type       string                  `json:"type"`
	Properties map[string]JsonProperty `json:"properties"`
}

type SchemaStruct struct {
	Schema     string `json:"schema"`
	SchemaType string `json:"schemaType"`
}

func ExtractSchemaIdFromResponse(response *http.Response) (int32, error) {
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(response.Body)
	if err != nil {
		return 0, err
	}
	responseBody := buf.String() // {"id":9}
	schemaId, err := strconv.ParseInt(responseBody[6:len(responseBody)-1], 10, 32)
	if err != nil {
		return 0, err
	}
	return int32(schemaId), nil
}

func GetCAClient(caCertPath string) *http.Client {
	caCert, err := ioutil.ReadFile(caCertPath)
	if err != nil {
		log.Fatal(err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: caCertPool,
			},
		},
	}
	return client
}

func ConvertSchema(schemaType string, schemaBytes []byte) (string, error) {
	fmt.Println("converting for:", schemaType)
	var requestString string
	if schemaType == "AVRO" {
		schema := AvroSchema{}
		err := json.Unmarshal([]byte(schemaBytes), &schema)
		if err != nil {
			return "", err
		}
		requestString, err = ConvertAvroSchema(schema)
		if err != nil {
			return "", err
		}
	} else if schemaType == "JSON" {
		schema := JsonSchema{}
		err := json.Unmarshal([]byte(schemaBytes), &schema)
		if err != nil {
			return "", err
		}
		requestString, err = ConvertJsonSchema(schema)
		if err != nil {
			return "", err
		}
	} else {
		fmt.Println("Invalid schema type.")
	}
	return requestString, nil
}

func ConvertAvroSchema(schema AvroSchema) (string, error) {
	request := SchemaStruct{}
	request.SchemaType = "AVRO"
	request.Schema = `{ "type":"` + schema.Type + `", "name":"` + schema.SchemaName + `", "fields": [ `
	var fields []string
	for _, field := range schema.Fields {
		var components []string
		for key, value := range field {
			f := `"` + key + `":"` + value + `"`
			components = append(components, f)
		}
		fields = append(fields, "{"+strings.Join(components, ",")+"}")
	}
	request.Schema += strings.Join(fields, ",") + `]}`
	requestBytes, err := json.Marshal(request)
	if err != nil {
		return "", err
	}
	return string(requestBytes), nil
}

func ConvertJsonSchema(schema JsonSchema) (string, error) {
	request := SchemaStruct{}
	request.SchemaType = "JSON"
	request.Schema = `{"type":"` + schema.Type + `", "properties":{`
	var fields []string
	for name, property := range schema.Properties {
		var components []string
		for key, value := range property {
			f := `"` + key + `":"` + value + `"`
			components = append(components, f)
		}
		field := "{" + strings.Join(components, ",") + "}"
		fields = append(fields, `"`+name+`":`+field)
	}
	request.Schema += strings.Join(fields, ",") + `}}`
	requestBytes, err := json.Marshal(request)
	if err != nil {
		return "", err
	}
	return string(requestBytes), nil
}

func GetRegisterSchemaRequest(requestUrl, mdsToken, schemaType, schemaPath string) (*http.Request, error) {
	schemaBytes, err := ioutil.ReadFile(schemaPath)
	if err != nil {
		return nil, err
	}
	requestString, err := ConvertSchema(schemaType, schemaBytes)
	if err != nil {
		return nil, err
	}
	requestReader := strings.NewReader(requestString)

	req, err := http.NewRequest("POST", requestUrl, requestReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+mdsToken)
	req.Header.Set("Content-Type", "application/vnd.schemaregistry.v1+json")
	return req, nil
}

func GetAndWriteSchemaBySubject(srEndpoint, caCertPath, tempStorePath, subject, version, mdsToken string) error {
	requestUrl := srEndpoint + "/subjects/" + subject + "/versions/" + version
	return requestAndWriteSchema(requestUrl, caCertPath, tempStorePath, mdsToken)
}

func GetAndWriteSchemaById(srEndpoint, caCertPath, tempStorePath, schemaID, mdsToken string) error {
	requestUrl := srEndpoint + "/schemas/ids/" + schemaID
	return requestAndWriteSchema(requestUrl, caCertPath, tempStorePath, mdsToken)
}

func requestAndWriteSchema(requestUrl, caCertPath, tempStorePath, mdsToken string) error {
	req, err := http.NewRequest("GET", requestUrl, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+mdsToken)
	client := GetCAClient(caCertPath)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return err
	}
	responseBody := buf.String()

	schemaResponse := SchemaStruct{}
	err = json.Unmarshal([]byte(responseBody), &schemaResponse)
	if err != nil {
		return err
	}
	var schemaBytes []byte
	if schemaResponse.SchemaType == "JSON" {
		schema := JsonSchema{}
		err = json.Unmarshal([]byte(schemaResponse.Schema), &schema)
		if err != nil {
			return err
		}
		schemaBytes, _ = json.MarshalIndent(schema, "", " ")
	} else if schemaResponse.SchemaType == "AVRO" {
		schema := AvroSchema{}
		err = json.Unmarshal([]byte(schemaResponse.Schema), &schema)
		if err != nil {
			return err
		}
		schemaBytes, _ = json.MarshalIndent(schema, "", " ")
	}
	return ioutil.WriteFile(tempStorePath, schemaBytes, 0644)
}
