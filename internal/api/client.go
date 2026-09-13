package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const aptEndpoint = "https://api.odcloud.kr/api/ApplyhomeInfoDetailSvc/v1/getAPTLttotPblancDetail"

type Client struct {
	ServiceKey string
	HTTPClient *http.Client
}

func NewClient(serviceKey string) *Client {
	return &Client{
		ServiceKey: serviceKey,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
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
	const perPage = 500
	var all []AptListing

	page := 1
	for {
		resp, err := c.fetchPage(page, perPage)
		if err != nil {
			return nil, fmt.Errorf("fetch page %d: %w", page, err)
		}
		all = append(all, resp.Data...)

		if page*perPage >= resp.TotalCount || len(resp.Data) == 0 {
			break
		}
		page++
	}
	return all, nil
}

func (c *Client) fetchPage(page, perPage int) (*apiResponse, error) {
	body, err := c.fetchPageRaw(page, perPage)
	if err != nil {
		return nil, err
	}

	var out apiResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &out, nil
}

// FetchPageRaw returns the raw, undecoded JSON response body for one page.
// It exists so callers (see cmd/inspect) can inspect the API's actual field
// names before trusting the AptListing struct tags to match them.
func (c *Client) FetchPageRaw(page, perPage int) ([]byte, error) {
	return c.fetchPageRaw(page, perPage)
}

func (c *Client) fetchPageRaw(page, perPage int) ([]byte, error) {
	q := url.Values{}
	q.Set("serviceKey", c.ServiceKey)
	q.Set("page", fmt.Sprintf("%d", page))
	q.Set("perPage", fmt.Sprintf("%d", perPage))

	reqURL := aptEndpoint + "?" + q.Encode()

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
