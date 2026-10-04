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
