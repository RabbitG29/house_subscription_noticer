package api

import (
	"encoding/json"
	"testing"
)

// 실제 getAPTLttotPblancMdl 응답(2026820011 공고)에서 가져온 형태입니다.
// 세대수는 숫자, 면적/분양가는 문자열이라는 타입 차이를 고정해 둡니다.
const sampleModelResponse = `{
  "currentCount": 1, "matchCount": 1, "page": 1, "perPage": 100, "totalCount": 14782,
  "data": [{
    "ETC_HSHLDCO": 0, "HOUSE_MANAGE_NO": "2026820011", "HOUSE_TY": "056.0000O",
    "INSTT_RECOMEND_HSHLDCO": 0, "LFE_FRST_HSHLDCO": 5, "LTTOT_TOP_AMOUNT": "39358",
    "MNYCH_HSHLDCO": 0, "MODEL_NO": "02", "NWBB_HSHLDCO": 0, "NWWDS_HSHLDCO": 55,
    "OLD_PARNTS_SUPORT_HSHLDCO": 0, "PBLANC_NO": "2026820011", "SPSPLY_HSHLDCO": 60,
    "SUPLY_AR": "85.5745", "SUPLY_HSHLDCO": 0, "TRANSR_INSTT_ENFSN_HSHLDCO": 0, "YGMN_HSHLDCO": 0
  }]
}`

func TestDecodeModelResponse(t *testing.T) {
	var resp apiResponse[AptModel]
	if err := json.Unmarshal([]byte(sampleModelResponse), &resp); err != nil {
		t.Fatalf("decode 실패: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("len(Data) = %d, want 1", len(resp.Data))
	}
	m := resp.Data[0]
	if m.AnnouncementNo != "2026820011" || m.TopPrice != "39358" || m.SupplyArea != "85.5745" {
		t.Errorf("문자열 필드 불일치: %+v", m)
	}
	if m.SpecialUnits != 60 || m.NewlywedUnits != 55 || m.FirstLifeUnits != 5 {
		t.Errorf("세대수 필드 불일치: %+v", m)
	}
}

// 실제 getAPTLttotPblancCmpet 응답(2026000454 공고)에서 가져온 형태입니다.
// 접수건수/경쟁률이 숫자가 아닌 문자열이고, 경쟁률에 "(△1)" 같은 값이 온다는 점을 고정합니다.
const sampleCompetitionResponse = `{
  "currentCount": 2, "matchCount": 2, "page": 1, "perPage": 100, "totalCount": 55212,
  "data": [
    {"CMPET_RATE": "(△1)", "HOUSE_MANAGE_NO": "2026000454", "HOUSE_TY": "084.9890A", "MODEL_NO": "01",
     "PBLANC_NO": "2026000454", "REQ_CNT": "4", "RESIDE_SECD": "01", "RESIDE_SENM": "해당지역",
     "SUBSCRPT_RANK_CODE": 1, "SUPLY_HSHLDCO": 5},
    {"CMPET_RATE": "-", "HOUSE_TY": "084.9890A", "PBLANC_NO": "2026000454", "REQ_CNT": "2",
     "RESIDE_SENM": "기타지역", "SUBSCRPT_RANK_CODE": 1, "SUPLY_HSHLDCO": 5}
  ]
}`

func TestDecodeCompetitionResponse(t *testing.T) {
	var resp apiResponse[AptCompetition]
	if err := json.Unmarshal([]byte(sampleCompetitionResponse), &resp); err != nil {
		t.Fatalf("decode 실패: %v", err)
	}
	if len(resp.Data) != 2 {
		t.Fatalf("len(Data) = %d, want 2", len(resp.Data))
	}
	c := resp.Data[0]
	if c.AnnouncementNo != "2026000454" || c.Rank != 1 || c.ResideName != "해당지역" ||
		c.SupplyUnits != 5 || c.RequestCount != "4" || c.Rate != "(△1)" {
		t.Errorf("필드 불일치: %+v", c)
	}
}

// 실제 getAPTSpsplyReqstStus 응답(2025000205 공고)의 일부입니다. 접수건수가 전부 숫자입니다.
const sampleSpecialResponse = `{
  "currentCount": 1, "matchCount": 4, "page": 1, "perPage": 100, "totalCount": 12408,
  "data": [{
    "CRSPAREA_LFE_FRST_CNT": 23, "CRSPAREA_MNYCH_CNT": 5, "CRSPAREA_NWWDS_NMTW_CNT": 14,
    "CTPRVN_MNYCH_CNT": 1, "ETC_AREA_LFE_FRST_CNT": 52, "ETC_AREA_NWWDS_NMTW_CNT": 18,
    "HOUSE_TY": "061.9767A", "INSTT_RECOMEND_HSHLDCO": 1, "INSTT_RECOMEND_DCSN_CNT": 0,
    "LFE_FRST_HSHLDCO": 1, "MNYCH_HSHLDCO": 1, "NWWDS_NMTW_HSHLDCO": 1, "PBLANC_NO": "2025000205",
    "SPSPLY_HSHLDCO": 4, "SUBSCRPT_RESULT_NM": "청약접수 종료"
  }]
}`

func TestDecodeSpecialStatusResponse(t *testing.T) {
	var resp apiResponse[AptSpecialStatus]
	if err := json.Unmarshal([]byte(sampleSpecialResponse), &resp); err != nil {
		t.Fatalf("decode 실패: %v", err)
	}
	s := resp.Data[0]
	if s.TotalUnits != 4 || s.FirstLifeUnits != 1 || s.FirstLifeLocal != 23 || s.FirstLifeOther != 52 ||
		s.NewlywedLocal != 14 || s.NewlywedOther != 18 || s.MultiChildLocal != 5 || s.MultiChildState != 1 {
		t.Errorf("필드 불일치: %+v", s)
	}
}
