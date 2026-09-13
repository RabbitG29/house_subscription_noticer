// Command inspect calls the 청약홈 getAPTLttotPblancDetail API with the
// configured service key and prints the raw, undecoded JSON response.
//
// Use this before trusting internal/api/types.go's struct tags: run it once
// with a real DATA_GO_KR_SERVICE_KEY and compare the printed field names
// (especially anything address-related) against AptListing's `json` tags.
//
//	DATA_GO_KR_SERVICE_KEY=xxx go run ./cmd/inspect
package main

import (
	"bytes"
	"encoding/json"
	"log"
	"os"

	"apply-alert-bot/internal/api"
)

func main() {
	serviceKey := os.Getenv("DATA_GO_KR_SERVICE_KEY")
	if serviceKey == "" {
		log.Fatal("DATA_GO_KR_SERVICE_KEY 환경변수가 필요합니다.")
	}

	client := api.NewClient(serviceKey)

	raw, err := client.FetchPageRaw(1, 3)
	if err != nil {
		log.Fatalf("API 호출 실패: %v", err)
	}

	var pretty bytes.Buffer
	if err := json.Indent(&pretty, raw, "", "  "); err != nil {
		log.Fatalf("응답이 JSON이 아닙니다(원문 출력): %v\n%s", err, raw)
	}

	os.Stdout.Write(pretty.Bytes())
	os.Stdout.Write([]byte("\n"))
}
