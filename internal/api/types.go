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
	HouseTypeName      string `json:"HOUSE_SECD_NM"`         // 주택 대분류 (APT/오피스텔 등)
	HouseDetailType    string `json:"HOUSE_DTL_SECD_NM"`     // 분양 유형 (민영/국민 등)
	NoticeDate         string `json:"RCRIT_PBLANC_DE"`       // 모집공고일 (YYYY-MM-DD)
	TotalSupplyUnits   int    `json:"TOT_SUPLY_HSHLDCO"`     // 총 공급세대수
	ReceiptStart       string `json:"RCEPT_BGNDE"`           // 청약접수 시작일
	ReceiptEnd         string `json:"RCEPT_ENDDE"`           // 청약접수 종료일
	WinnerAnnounceDate string `json:"PRZWNER_PRESNATN_DE"`   // 당첨자발표일
	HomepageURL        string `json:"HMPG_ADRES"`            // 분양 홈페이지 주소
}

// AptModel represents one row from the "APT 분양정보 주택형별 상세조회" API
// (getAPTLttotPblancMdl) — 공고 하나에 딸린 주택형(평형)별 공급 정보입니다.
// 실제 응답으로 필드명/타입을 확인했습니다: 세대수는 JSON 숫자,
// 면적(SUPLY_AR)과 분양가(LTTOT_TOP_AMOUNT)는 JSON 문자열로 내려옵니다.
type AptModel struct {
	AnnouncementNo string `json:"PBLANC_NO"`        // 공고번호 (AptListing.AnnouncementNo와 조인)
	ModelNo        string `json:"MODEL_NO"`         // 주택형 순번 ("01", "02", ...)
	HouseType      string `json:"HOUSE_TY"`         // 주택형 (예: "084.9500A")
	SupplyArea     string `json:"SUPLY_AR"`         // 공급면적(㎡), 문자열
	TopPrice       string `json:"LTTOT_TOP_AMOUNT"` // 공급금액(분양최고가), 단위: 만원, 문자열

	GeneralUnits int `json:"SUPLY_HSHLDCO"`  // 일반공급 세대수
	SpecialUnits int `json:"SPSPLY_HSHLDCO"` // 특별공급 세대수 합계

	// 특별공급 유형별 세대수 (합이 SpecialUnits와 같도록 내려옴)
	MultiChildUnits    int `json:"MNYCH_HSHLDCO"`              // 다자녀가구
	NewlywedUnits      int `json:"NWWDS_HSHLDCO"`              // 신혼부부
	FirstLifeUnits     int `json:"LFE_FRST_HSHLDCO"`           // 생애최초
	OldParentsUnits    int `json:"OLD_PARNTS_SUPORT_HSHLDCO"`  // 노부모부양
	InstitutionUnits   int `json:"INSTT_RECOMEND_HSHLDCO"`     // 기관추천
	YoungUnits         int `json:"YGMN_HSHLDCO"`               // 청년
	NewbornUnits       int `json:"NWBB_HSHLDCO"`               // 신생아
	TransferAgencyUnit int `json:"TRANSR_INSTT_ENFSN_HSHLDCO"` // 이전기관종사자
	EtcUnits           int `json:"ETC_HSHLDCO"`                // 기타
}

// apiResponse는 두 엔드포인트가 공유하는 odcloud 응답 봉투입니다.
// Data의 원소 타입만 다르므로 제네릭으로 한 번만 정의합니다.
type apiResponse[T any] struct {
	CurrentCount int `json:"currentCount"`
	Data         []T `json:"data"`
	MatchCount   int `json:"matchCount"`
	Page         int `json:"page"`
	PerPage      int `json:"perPage"`
	TotalCount   int `json:"totalCount"`
}
