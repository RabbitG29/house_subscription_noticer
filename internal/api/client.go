package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const detailBase = "https://api.odcloud.kr/api/ApplyhomeInfoDetailSvc/v1/"

const (
	aptEndpoint      = detailBase + "getAPTLttotPblancDetail"
	aptModelEndpoint = detailBase + "getAPTLttotPblancMdl"
)

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
		resp, err := fetchPage[AptListing](c, aptEndpoint, nil, page, perPage)
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

// FetchModelsByAnnouncement는 공고번호 하나에 딸린 주택형별 정보
// (getAPTLttotPblancMdl)를 가져옵니다. 이 엔드포인트는 전체가 1만 건이 넘기
// 때문에 전체를 받지 않고 서버 쪽 필터 cond[PBLANC_NO::EQ]로 해당 공고만
// 조회합니다. 한 공고의 주택형은 많아야 수십 개라 perPage=100 한 번이면
// 충분하지만, 안전하게 페이지네이션도 처리합니다.
func (c *Client) FetchModelsByAnnouncement(announcementNo string) ([]AptModel, error) {
	const perPage = 100
	cond := url.Values{}
	cond.Set("cond[PBLANC_NO::EQ]", announcementNo)

	var all []AptModel
	page := 1
	for {
		resp, err := fetchPage[AptModel](c, aptModelEndpoint, cond, page, perPage)
		if err != nil {
			return nil, fmt.Errorf("fetch models of %s page %d: %w", announcementNo, page, err)
		}
		all = append(all, resp.Data...)

		if page*perPage >= resp.MatchCount || len(resp.Data) == 0 {
			break
		}
		page++
	}
	return all, nil
}

// fetchPage는 메서드가 될 수 없습니다 — Go는 메서드에 타입 파라미터(제네릭)를
// 붙이는 것을 허용하지 않아서, Client를 첫 인자로 받는 일반 함수로 둡니다.
func fetchPage[T any](c *Client, endpoint string, extra url.Values, page, perPage int) (*apiResponse[T], error) {
	body, err := c.fetchRaw(endpoint, extra, page, perPage)
	if err != nil {
		return nil, err
	}

	var out apiResponse[T]
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &out, nil
}

// FetchPageRaw returns the raw, undecoded JSON response body for one page.
// It exists so callers (see cmd/inspect) can inspect the API's actual field
// names before trusting the AptListing struct tags to match them.
func (c *Client) FetchPageRaw(page, perPage int) ([]byte, error) {
	return c.fetchRaw(aptEndpoint, nil, page, perPage)
}

func (c *Client) fetchRaw(endpoint string, extra url.Values, page, perPage int) ([]byte, error) {
	q := url.Values{}
	for k, vs := range extra {
		for _, v := range vs {
			q.Add(k, v)
		}
	}
	q.Set("serviceKey", c.ServiceKey)
	q.Set("page", fmt.Sprintf("%d", page))
	q.Set("perPage", fmt.Sprintf("%d", perPage))

	reqURL := endpoint + "?" + q.Encode()

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
