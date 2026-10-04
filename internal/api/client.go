package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	defaultBaseURL = "https://api.odcloud.kr/api/ApplyhomeInfoDetailSvc/v1/"

	opAptDetail = "getAPTLttotPblancDetail" // 분양정보 상세조회
	opAptModel  = "getAPTLttotPblancMdl"    // 분양정보 주택형별 상세조회

	// OpDetail, OpModel은 FetchRaw에 넘길 수 있는 오퍼레이션명입니다.
	OpDetail = opAptDetail
	OpModel  = opAptModel
)

type Client struct {
	ServiceKey string
	HTTPClient *http.Client
	BaseURL    string // 테스트에서 httptest 서버로 바꿔 끼우기 위한 용도
}

func NewClient(serviceKey string) *Client {
	return &Client{
		ServiceKey: serviceKey,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
		BaseURL:    defaultBaseURL,
	}
}

// FetchAllAptListings paginates through the API and returns every listing
// currently published. The dataset is small (수천 건 단위)이므로 매번
// 전체를 받아서 이미 본 공고와 diff하는 방식이, "마지막 조회 이후"를
// 서버에 물어보는 것보다 훨씬 단순합니다.
//
// perPage=1000으로 실측했을 때 서버 응답이 비정상적으로 느려져(30초+)
// 타임아웃이 반복 발생했습니다. perPage=500은 매번 10초 이내에 안정적으로
// 끝나서 이 값으로 낮췄습니다 — odcloud.kr 쪽의 큰 페이지 처리 이슈로
// 보이며, 우리 쪽에서 더 줄일 필요가 생기면 이 상수만 바꾸면 됩니다.
func (c *Client) FetchAllAptListings() ([]AptListing, error) {
	return fetchAll[AptListing](c, opAptDetail, 500, nil)
}

// FetchModels returns the 주택형별 정보(특별공급 세대수, 분양가 등) of one
// announcement. 서버 측 조건 검색(cond[...::EQ])으로 해당 공고만 받아옵니다.
func (c *Client) FetchModels(houseManageNo, announcementNo string) ([]AptModel, error) {
	if houseManageNo == "" {
		return nil, fmt.Errorf("주택관리번호(HOUSE_MANAGE_NO)가 비어 있습니다")
	}
	conds := url.Values{}
	conds.Set("cond[HOUSE_MANAGE_NO::EQ]", houseManageNo)
	if announcementNo != "" {
		conds.Set("cond[PBLANC_NO::EQ]", announcementNo)
	}
	return fetchAll[AptModel](c, opAptModel, 100, conds)
}

// fetchAll은 메서드에 타입 파라미터를 쓸 수 없어 패키지 함수로 둡니다.
func fetchAll[T any](c *Client, op string, perPage int, conds url.Values) ([]T, error) {
	var all []T

	for page := 1; ; page++ {
		body, err := c.fetchRaw(op, page, perPage, conds)
		if err != nil {
			return nil, fmt.Errorf("fetch %s page %d: %w", op, page, err)
		}

		var resp apiResponse[T]
		if err := json.Unmarshal(body, &resp); err != nil {
			return nil, fmt.Errorf("decode %s page %d: %w", op, page, err)
		}
		all = append(all, resp.Data...)

		if page*perPage >= resp.TotalCount || len(resp.Data) == 0 {
			return all, nil
		}
	}
}

// FetchPageRaw returns the raw, undecoded JSON response body for one page
// of getAPTLttotPblancDetail.
// It exists so callers (see cmd/inspect) can inspect the API's actual field
// names before trusting the AptListing struct tags to match them.
func (c *Client) FetchPageRaw(page, perPage int) ([]byte, error) {
	return c.fetchRaw(opAptDetail, page, perPage, nil)
}

// FetchRaw is FetchPageRaw for an arbitrary operation (OpDetail, OpModel)
// with optional cond[...] query parameters.
func (c *Client) FetchRaw(op string, page, perPage int, conds url.Values) ([]byte, error) {
	return c.fetchRaw(op, page, perPage, conds)
}

func (c *Client) fetchRaw(op string, page, perPage int, conds url.Values) ([]byte, error) {
	q := url.Values{}
	for k, vs := range conds {
		for _, v := range vs {
			q.Add(k, v)
		}
	}
	q.Set("serviceKey", c.ServiceKey)
	q.Set("page", strconv.Itoa(page))
	q.Set("perPage", strconv.Itoa(perPage))

	reqURL := c.BaseURL + op + "?" + q.Encode()

	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", res.StatusCode, string(body))
	}

	return body, nil
}
