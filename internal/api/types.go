package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// AptListing represents one row from the 한국부동산원 청약홈
// "APT 분양정보 상세조회" API (getAPTLttotPblancDetail).
//
// 필드명은 승인된 서비스키로 직접 호출해 실제 응답과 대조하여
// 검증했습니다(cmd/inspect). TOT_SUPLY_HSHLDCO는 문서상 추정과 달리
// 문자열이 아닌 JSON 숫자로 내려와 디코딩이 실패하던 것을 바로잡았고,
// HSSPLY_ADRES(공급위치 상세주소)가 실제로 존재함을 확인해 추가했습니다.
type AptListing struct {
	HouseManageNo      FlexString `json:"HOUSE_MANAGE_NO"`       // 주택관리번호 (주택형별 API 조회 키)
	AnnouncementNo     string     `json:"PBLANC_NO"`             // 공고번호 (고유 ID로 사용)
	HouseName          string     `json:"HOUSE_NM"`              // 주택명
	RegionName         string     `json:"SUBSCRPT_AREA_CODE_NM"` // 공급지역명 (예: "경기") — 시/도 단위
	SupplyAddress      string     `json:"HSSPLY_ADRES"`          // 공급위치 상세주소 (예: "경기도 수원시 권선구 ...") — 시/구 단위 필터링에 사용
	HouseTypeName      string     `json:"HOUSE_SECD_NM"`         // 주택 대분류 (APT/오피스텔 등)
	HouseDetailType    string     `json:"HOUSE_DTL_SECD_NM"`     // 분양 유형 (민영/국민 등)
	NoticeDate         string     `json:"RCRIT_PBLANC_DE"`       // 모집공고일 (YYYY-MM-DD)
	TotalSupplyUnits   int        `json:"TOT_SUPLY_HSHLDCO"`     // 총 공급세대수
	ReceiptStart       string     `json:"RCEPT_BGNDE"`           // 청약접수 시작일
	ReceiptEnd         string     `json:"RCEPT_ENDDE"`           // 청약접수 종료일
	WinnerAnnounceDate string     `json:"PRZWNER_PRESNATN_DE"`   // 당첨자발표일
	HomepageURL        string     `json:"HMPG_ADRES"`            // 분양 홈페이지 주소
}

// AptModel represents one row (= 주택형 하나) from the 청약홈
// "APT 분양정보 주택형별 상세조회" API (getAPTLttotPblancMdl).
//
// 주의: 아래 필드명은 공공데이터포털 문서 기준 후보이며, 서비스키로 직접
// 호출한 실제 응답(cmd/inspect -op=model)과 아직 대조하지 않았습니다.
// 값이 숫자/문자열 어느 쪽으로 내려와도 디코딩이 깨지지 않도록 Number,
// FlexString 타입을 사용합니다.
type AptModel struct {
	HouseManageNo  FlexString `json:"HOUSE_MANAGE_NO"` // 주택관리번호
	AnnouncementNo FlexString `json:"PBLANC_NO"`       // 공고번호
	ModelNo        FlexString `json:"MODEL_NO"`        // 모델번호
	HouseType      string     `json:"HOUSE_TY"`        // 주택형 (예: "084.9500A")
	SupplyArea     Number     `json:"SUPLY_AR"`        // 공급면적(㎡)
	GeneralUnits   Number     `json:"SUPLY_HSHLDCO"`   // 일반공급 세대수
	SpecialUnits   Number     `json:"SPSPLY_HSHLDCO"`  // 특별공급 세대수 합계

	MultiChildUnits    Number `json:"MNYCH_HSHLDCO"`              // 다자녀가구
	NewlywedUnits      Number `json:"NWWDS_HSHLDCO"`              // 신혼부부
	FirstLifeUnits     Number `json:"LFE_FRST_HSHLDCO"`           // 생애최초
	OldParentsUnits    Number `json:"OLD_PARNTS_SUPORT_HSHLDCO"`  // 노부모부양
	InstRecommendUnits Number `json:"INSTT_RECOMEND_HSHLDCO"`     // 기관추천
	TransferUnits      Number `json:"TRANSR_INSTT_ENFSN_HSHLDCO"` // 이전기관
	YouthUnits         Number `json:"YGMN_HSHLDCO"`               // 청년
	NewbornUnits       Number `json:"NWBB_HSHLDCO"`               // 신생아
	EtcUnits           Number `json:"ETC_HSHLDCO"`                // 기타

	TopPrice Number `json:"LTTOT_TOP_AMOUNT"` // 분양최고금액 (단위: 만원으로 가정)
}

// SpecialUnit is one 특별공급 category with its household count.
type SpecialUnit struct {
	Name  string
	Units int
}

// SpecialBreakdown returns the 특별공급 categories with a non-zero household
// count, in a stable display order.
func (m AptModel) SpecialBreakdown() []SpecialUnit {
	all := []SpecialUnit{
		{"다자녀", m.MultiChildUnits.Int()},
		{"신혼부부", m.NewlywedUnits.Int()},
		{"생애최초", m.FirstLifeUnits.Int()},
		{"노부모부양", m.OldParentsUnits.Int()},
		{"기관추천", m.InstRecommendUnits.Int()},
		{"이전기관", m.TransferUnits.Int()},
		{"청년", m.YouthUnits.Int()},
		{"신생아", m.NewbornUnits.Int()},
		{"기타", m.EtcUnits.Int()},
	}
	out := all[:0]
	for _, u := range all {
		if u.Units > 0 {
			out = append(out, u)
		}
	}
	return out
}

// Number decodes a JSON number, a numeric string ("1,234", " 84.95 ") or
// null/"" (→ 0). 공공데이터 API는 같은 필드도 숫자/문자열이 섞여 내려오는
// 경우가 있어(TOT_SUPLY_HSHLDCO 사례) 관대하게 받습니다.
type Number float64

func (n *Number) UnmarshalJSON(b []byte) error {
	s := strings.TrimSpace(string(b))
	if s == "null" {
		*n = 0
		return nil
	}
	s = strings.Trim(s, `"`)
	s = strings.ReplaceAll(strings.TrimSpace(s), ",", "")
	if s == "" {
		*n = 0
		return nil
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return fmt.Errorf("숫자로 해석할 수 없는 값 %s: %w", b, err)
	}
	*n = Number(f)
	return nil
}

// Int returns the value rounded to the nearest integer.
func (n Number) Int() int {
	if n < 0 {
		return int(n - 0.5)
	}
	return int(n + 0.5)
}

// FlexString decodes a JSON string or number into a string (null → "").
type FlexString string

func (s *FlexString) UnmarshalJSON(b []byte) error {
	b = bytes.TrimSpace(b)
	if string(b) == "null" {
		*s = ""
		return nil
	}
	if len(b) > 0 && b[0] == '"' {
		var v string
		if err := json.Unmarshal(b, &v); err != nil {
			return err
		}
		*s = FlexString(strings.TrimSpace(v))
		return nil
	}
	*s = FlexString(b) // 숫자 리터럴은 원문 그대로 사용
	return nil
}

// apiResponse is the odcloud.kr paging envelope shared by every operation.
type apiResponse[T any] struct {
	CurrentCount int `json:"currentCount"`
	Data         []T `json:"data"`
	MatchCount   int `json:"matchCount"`
	Page         int `json:"page"`
	PerPage      int `json:"perPage"`
	TotalCount   int `json:"totalCount"`
}
