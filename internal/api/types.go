package api

// AptListing represents one row from the 한국부동산원 청약홈
// "APT 분양정보 상세조회" API (getAPTLttotPblancDetail).
//
// 필드명은 승인된 서비스키로 직접 호출해 실제 응답과 대조하여
// 검증했습니다(cmd/inspect). TOT_SUPLY_HSHLDCO는 문서상 추정과 달리
// 문자열이 아닌 JSON 숫자로 내려와 디코딩이 실패하던 것을 바로잡았고,
// HSSPLY_ADRES(공급위치 상세주소)가 실제로 존재함을 확인해 추가했습니다.
type AptListing struct {
	AnnouncementNo     string `json:"PBLANC_NO"`             // 공고번호 (고유 ID로 사용)
	HouseName          string `json:"HOUSE_NM"`              // 주택명
	RegionName         string `json:"SUBSCRPT_AREA_CODE_NM"` // 공급지역명 (예: "경기") — 시/도 단위
	SupplyAddress      string `json:"HSSPLY_ADRES"`          // 공급위치 상세주소 (예: "경기도 수원시 권선구 ...") — 시/구 단위 필터링에 사용
	HouseTypeName      string `json:"HOUSE_SECD_NM"`         // 주택 대분류 (APT/오피스텔 등 — "민영/국민" 구분은 HOUSE_DTL_SECD_NM)
	NoticeDate         string `json:"RCRIT_PBLANC_DE"`       // 모집공고일 (YYYY-MM-DD)
	TotalSupplyUnits   int    `json:"TOT_SUPLY_HSHLDCO"`     // 총 공급세대수
	ReceiptStart       string `json:"RCEPT_BGNDE"`           // 청약접수 시작일
	ReceiptEnd         string `json:"RCEPT_ENDDE"`           // 청약접수 종료일
	WinnerAnnounceDate string `json:"PRZWNER_PRESNATN_DE"`   // 당첨자발표일
	HomepageURL        string `json:"HMPG_ADRES"`            // 분양 홈페이지 주소
}

type apiResponse struct {
	CurrentCount int          `json:"currentCount"`
	Data         []AptListing `json:"data"`
	MatchCount   int          `json:"matchCount"`
	Page         int          `json:"page"`
	PerPage      int          `json:"perPage"`
	TotalCount   int          `json:"totalCount"`
}
