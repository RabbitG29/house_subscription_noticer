package main

import (
	"strings"
	"testing"

	"apply-alert-bot/internal/api"
)

func TestFormatPrice(t *testing.T) {
	tests := []struct{ in, want string }{
		{"36707", "3억 6,707만원"},
		{"50844", "5억 844만원"},
		{"100000", "10억원"},
		{"9500", "9,500만원"},
		{"500", "500만원"},
		{"", "정보 없음"},
		{"미정", "미정"},
	}
	for _, tc := range tests {
		if got := formatPrice(tc.in); got != tc.want {
			t.Errorf("formatPrice(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestFormatHouseType(t *testing.T) {
	tests := []struct{ in, want string }{
		{"084.9500A", "84.95A"},
		{"059.9900", "59.99"},
		{"055.0000O", "55O"},
		{"이상한값", "이상한값"},
	}
	for _, tc := range tests {
		if got := formatHouseType(tc.in); got != tc.want {
			t.Errorf("formatHouseType(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestBuildBody(t *testing.T) {
	l := api.AptListing{HouseName: "테스트단지", SupplyAddress: "경기도 수원시 영통구", TotalSupplyUnits: 100}

	t.Run("주택형 정보가 없으면 섹션을 생략", func(t *testing.T) {
		body := buildBody(l, nil)
		if strings.Contains(body, "주택형별") {
			t.Errorf("주택형 섹션이 없어야 함:\n%s", body)
		}
	})

	t.Run("주택형별 분양가와 특공 세대수 포함", func(t *testing.T) {
		models := []api.AptModel{{
			HouseType: "084.9500A", SupplyArea: "110.5", TopPrice: "85000",
			GeneralUnits: 40, SpecialUnits: 30, NewlywedUnits: 20, FirstLifeUnits: 10,
		}}
		body := buildBody(l, models)
		for _, want := range []string{
			"84.95A", "8억 5,000만원", "일반공급: 40세대",
			"특별공급: 30세대 (신혼부부 20, 생애최초 10)",
		} {
			if !strings.Contains(body, want) {
				t.Errorf("본문에 %q 없음:\n%s", want, body)
			}
		}
	})
}

func TestFormatRate(t *testing.T) {
	tests := []struct{ in, want string }{
		{"7.28", "7.28:1"},
		{"(△1)", "1세대 미달"},
		{"-", "경쟁률 미산정"},
		{"", "경쟁률 미산정"},
		{"이상한값", "이상한값"},
	}
	for _, tc := range tests {
		if got := formatRate(tc.in); got != tc.want {
			t.Errorf("formatRate(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestBuildCompetitionBody(t *testing.T) {
	l := api.AptListing{HouseName: "테스트단지", SupplyAddress: "경기도 수원시 영통구", ReceiptEnd: "2026-10-14"}
	comps := []api.AptCompetition{
		{HouseType: "084.9243A", Rank: 1, ResideName: "해당지역", SupplyUnits: 175, RequestCount: "1274", Rate: "7.28"},
		{HouseType: "084.9243A", Rank: 1, ResideName: "기타지역", SupplyUnits: 175, RequestCount: "854", Rate: "-"},
		{HouseType: "084.9243A", Rank: 2, ResideName: "해당지역", SupplyUnits: 175, RequestCount: "0", Rate: "-"},
		{HouseType: "059.9900B", Rank: 1, ResideName: "해당지역", SupplyUnits: 5, RequestCount: "4", Rate: "(△1)"},
	}

	t.Run("일반공급 경쟁률만", func(t *testing.T) {
		body := buildCompetitionBody(l, comps, nil)
		for _, want := range []string{
			"■ 84.9243A (일반공급 175세대)", "1순위 해당지역: 접수 1274건 · 7.28:1",
			"1순위 기타지역: 접수 854건 · 경쟁률 미산정", "■ 59.99B (일반공급 5세대)", "1세대 미달",
		} {
			if !strings.Contains(body, want) {
				t.Errorf("본문에 %q 없음:\n%s", want, body)
			}
		}
		if strings.Contains(body, "2순위") {
			t.Errorf("접수 0건인 2순위 행은 빠져야 함:\n%s", body)
		}
		if strings.Contains(body, "특별공급") {
			t.Errorf("특별공급 섹션이 없어야 함:\n%s", body)
		}
	})

	t.Run("특별공급 신청현황 포함", func(t *testing.T) {
		specials := []api.AptSpecialStatus{{
			HouseType: "061.9767A", TotalUnits: 4,
			FirstLifeUnits: 1, FirstLifeLocal: 23, FirstLifeState: 0, FirstLifeOther: 52,
			NewlywedUnits: 2, NewlywedLocal: 3,
		}}
		body := buildCompetitionBody(l, comps, specials)
		for _, want := range []string{
			"[특별공급 신청현황]", "■ 61.9767A (특별공급 4세대)",
			"생애최초 1세대: 접수 75건 (해당지역 23, 해당 시·도 0, 기타 52) · 약 75.0:1",
			"신혼부부 2세대: 접수 3건",
		} {
			if !strings.Contains(body, want) {
				t.Errorf("본문에 %q 없음:\n%s", want, body)
			}
		}
		if strings.Contains(body, "다자녀") {
			t.Errorf("세대수·접수가 0인 유형은 빠져야 함:\n%s", body)
		}
	})
}
