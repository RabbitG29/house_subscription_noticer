// Command inspect calls the 청약홈 getAPTLttotPblancDetail API with the
// configured service key and prints the raw, undecoded JSON response.
//
// Use this before trusting internal/api/types.go's struct tags: run it once
// with a real DATA_GO_KR_SERVICE_KEY and compare the printed field names
// (especially anything address-related) against AptListing's `json` tags.
//
//	DATA_GO_KR_SERVICE_KEY=xxx go run ./cmd/inspect
//
// 주택형별 상세조회(getAPTLttotPblancMdl)는 -op=model과 주택관리번호로 확인합니다.
//
//	DATA_GO_KR_SERVICE_KEY=xxx go run ./cmd/inspect -op=model -house=<HOUSE_MANAGE_NO>
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"log"
	"net/url"
	"os"

	"apply-alert-bot/internal/api"
)

func main() {
	serviceKey := os.Getenv("DATA_GO_KR_SERVICE_KEY")
	if serviceKey == "" {
		log.Fatal("DATA_GO_KR_SERVICE_KEY 환경변수가 필요합니다.")
	}

	op := flag.String("op", "detail", "조회할 API: detail(분양정보 상세) | model(주택형별 상세)")
	house := flag.String("house", "", "op=model일 때 조회할 HOUSE_MANAGE_NO")
	flag.Parse()

	client := api.NewClient(serviceKey)

	var raw []byte
	var err error
	switch *op {
	case "detail":
		raw, err = client.FetchPageRaw(1, 3)
	case "model":
		if *house == "" {
			log.Fatal("-op=model에는 -house=<HOUSE_MANAGE_NO>가 필요합니다.")
		}
		conds := url.Values{}
		conds.Set("cond[HOUSE_MANAGE_NO::EQ]", *house)
		raw, err = client.FetchRaw(api.OpModel, 1, 20, conds)
	default:
		log.Fatalf("알 수 없는 -op 값: %q", *op)
	}
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
