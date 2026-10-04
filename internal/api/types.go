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

// AptCompetition represents one row from the "APT 분양정보 경쟁률" API
// (getAPTLttotPblancCmpet) — 주택형 × 순위 × 거주지역 조합마다 한 행입니다.
// 실제 응답으로 확인한 사항: 접수건수(REQ_CNT)와 경쟁률(CMPET_RATE)은 JSON
// 문자열이고, 경쟁률은 "7.28" 같은 숫자 외에 "-"(상위 지역에서 마감되어 산정
// 안 됨)나 "(△1)"(1세대 미달) 값도 오므로 숫자로 파싱하지 않고 문자열로 둡니다.
type AptCompetition struct {
	AnnouncementNo string `json:"PBLANC_NO"`          // 공고번호
	HouseType      string `json:"HOUSE_TY"`           // 주택형 (예: "084.9243A")
	Rank           int    `json:"SUBSCRPT_RANK_CODE"` // 청약 순위 (1, 2)
	ResideName     string `json:"RESIDE_SENM"`        // 거주지역 구분 ("해당지역"/"기타지역" 등)
	SupplyUnits    int    `json:"SUPLY_HSHLDCO"`      // 공급세대수
	RequestCount   string `json:"REQ_CNT"`            // 접수건수, 문자열
	Rate           string `json:"CMPET_RATE"`         // 경쟁률 — "7.28", "-", "(△1)"
}

// AptSpecialStatus represents one row from the "APT 특별공급 신청현황" API
// (getAPTSpsplyReqstStus) — 주택형마다 한 행이며, 특별공급 유형별로
// 공급세대수(*_HSHLDCO)와 접수건수(*_CNT)가 가로로 펼쳐져 있습니다.
// 접수건수는 거주지역별로 해당지역(CRSPAREA_) / 해당 시·도(CTPRVN_) /
// 기타지역(ETC_AREA_)으로 나뉩니다. 모두 JSON 숫자입니다.
type AptSpecialStatus struct {
	AnnouncementNo string `json:"PBLANC_NO"`
	HouseType      string `json:"HOUSE_TY"`
	TotalUnits     int    `json:"SPSPLY_HSHLDCO"` // 특별공급 세대수 합계

	MultiChildUnits int `json:"MNYCH_HSHLDCO"` // 다자녀가구
	MultiChildLocal int `json:"CRSPAREA_MNYCH_CNT"`
	MultiChildState int `json:"CTPRVN_MNYCH_CNT"`
	MultiChildOther int `json:"ETC_AREA_MNYCH_CNT"`

	NewlywedUnits int `json:"NWWDS_NMTW_HSHLDCO"` // 신혼부부
	NewlywedLocal int `json:"CRSPAREA_NWWDS_NMTW_CNT"`
	NewlywedState int `json:"CTPRVN_NWWDS_NMTW_CNT"`
	NewlywedOther int `json:"ETC_AREA_NWWDS_NMTW_CNT"`

	FirstLifeUnits int `json:"LFE_FRST_HSHLDCO"` // 생애최초
	FirstLifeLocal int `json:"CRSPAREA_LFE_FRST_CNT"`
	FirstLifeState int `json:"CTPRVN_LFE_FRST_CNT"`
	FirstLifeOther int `json:"ETC_AREA_LFE_FRST_CNT"`

	OldParentsUnits int `json:"OLD_PARNTS_SUPORT_HSHLDCO"` // 노부모부양
	OldParentsLocal int `json:"CRSPAREA_OPS_CNT"`
	OldParentsState int `json:"CTPRVN_OPS_CNT"`
	OldParentsOther int `json:"ETC_AREA_OPS_CNT"`

	YoungUnits int `json:"YGMN_HSHLDCO"` // 청년
	YoungLocal int `json:"CRSPAREA_YGMN_CNT"`
	YoungState int `json:"CTPRVN_YGMN_CNT"`
	YoungOther int `json:"ETC_AREA_YGMN_CNT"`

	NewbornUnits int `json:"NWBB_NWBBSHR_HSHLDCO"` // 신생아
	NewbornLocal int `json:"CRSPAREA_NWBB_NWBBSHR_CNT"`
	NewbornState int `json:"CTPRVN_NWBB_NWBBSHR_CNT"`
	NewbornOther int `json:"ETC_AREA_NWBB_NWBBSHR_CNT"`

	// 기관추천·이전기관은 거주지역 구분 없이 접수건수만 내려옵니다.
	InstitutionUnits    int `json:"INSTT_RECOMEND_HSHLDCO"`
	InstitutionDecided  int `json:"INSTT_RECOMEND_DCSN_CNT"`   // 결정
	InstitutionPrepared int `json:"INSTT_RECOMEND_PREPAR_CNT"` // 준비
	TransferUnits       int `json:"TRANSR_INSTT_ENFSN_HSHLDCO"`
	TransferCount       int `json:"TRANSR_INSTT_ENFSN_CNT"`
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
