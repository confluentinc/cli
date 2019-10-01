package main

import (
	"context"
	"fmt"

	"github.com/confluentinc/ccloud-sdk-go"
	"github.com/confluentinc/ccloudapis/kafka/v1"

	"github.com/confluentinc/cli/internal/pkg/log"
)

func main() {
	token := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJvcmdhbml6YXRpb25JZCI6NDYzLCJ1c2VySWQiOjY3MCwiZXhwIjoxNTY2OTIyODEyLCJqdGkiOiIxZmQ3Yzk4Yi1iMDhkLTRlMGYtYTQzZi1mNWIzNTQxOWY4OGUiLCJpYXQiOjE1NjY5MTkyMTIsImlzcyI6IkNvbmZsdWVudCIsInN1YiI6IjY3MCJ9.cSOFyyfrvDbDRSxV4NSV5G357HsA10dU-xJv3XThT5_sq7CRharV39ydVu26zeJ3MiDvxeoMRWsZdklCYHOuhKH3YpNFrhNqb59sW9E8PaEd6JZyTaDsfZ_jge-wKCSLlBVvPPn44MHwSsXXNcBPMP0momMs7J6-1j6pTZyRCRQ1Go73RH3ec-gWSocpYEON5OTSn-RPG37srLZwuFAemU6_fbFqwTmdo4C5nvDNSlUQqRhxV89KwfHkAV0qVf_BznEY6K3p1-qrh_kSngLjS7_gmqVXQKD08XwsLLzJ1AmD90LtJrECTyBsjaP_GQZEAoJeFtJvxjhzy45r2Hl9Cw"
	userAgent := "Confluent-Cloud-CLI/v0.0.0 (https://confluent.cloud; support@confluent.io)"
	url := "https://confluent.cloud"
	accountID := "t463"

	client := ccloud.NewClientWithJWT(context.Background(), token, &ccloud.Params{
		BaseURL: url, Logger: log.New(), UserAgent: userAgent,
	})
	topics, err := client.Kafka.ListTopics(context.Background(), &v1.KafkaCluster{
		AccountId: accountID,
		Id: "lkc-l77k2",
		ApiEndpoint: "https://pkac-4nvd3.us-east-1.aws.confluent.cloud",
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(topics)
}
