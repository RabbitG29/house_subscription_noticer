package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchModels(t *testing.T) {
	var gotPath, gotHouse, gotPblanc string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotHouse = r.URL.Query().Get("cond[HOUSE_MANAGE_NO::EQ]")
		gotPblanc = r.URL.Query().Get("cond[PBLANC_NO::EQ]")
		// 숫자/문자열/null이 섞여 있어도 디코딩되어야 합니다.
		w.Write([]byte(`{"currentCount":1,"totalCount":1,"page":1,"perPage":100,"matchCount":1,
			"data":[{"HOUSE_MANAGE_NO":2024000123,"PBLANC_NO":"2024000123","HOUSE_TY":"084.9500A",
			"SUPLY_AR":"84.9500","SUPLY_HSHLDCO":120,"SPSPLY_HSHLDCO":"60","NWWDS_HSHLDCO":30,
			"MNYCH_HSHLDCO":null,"LTTOT_TOP_AMOUNT":"84,500"}]}`))
	}))
	defer srv.Close()

	c := NewClient("key")
	c.BaseURL = srv.URL + "/"

	models, err := c.FetchModels("2024000123", "2024000123")
	if err != nil {
		t.Fatalf("FetchModels: %v", err)
	}
	if gotPath != "/"+OpModel || gotHouse != "2024000123" || gotPblanc != "2024000123" {
		t.Errorf("잘못된 요청: path=%q house=%q pblanc=%q", gotPath, gotHouse, gotPblanc)
	}
	if len(models) != 1 {
		t.Fatalf("len(models) = %d, want 1", len(models))
	}
	m := models[0]
	if m.GeneralUnits.Int() != 120 || m.SpecialUnits.Int() != 60 || m.TopPrice.Int() != 84500 {
		t.Errorf("디코딩 결과가 이상합니다: %+v", m)
	}
	if sb := m.SpecialBreakdown(); len(sb) != 1 || sb[0].Name != "신혼부부" || sb[0].Units != 30 {
		t.Errorf("SpecialBreakdown = %+v", sb)
	}
}

func TestFetchModelsRequiresHouseManageNo(t *testing.T) {
	if _, err := NewClient("key").FetchModels("", "x"); err == nil {
		t.Error("빈 주택관리번호는 에러여야 합니다")
	}
}
