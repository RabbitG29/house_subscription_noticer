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
