package main

import (
	"errors"
	"strings"
	"testing"

	"apply-alert-bot/internal/api"
)

func TestFormatPrice(t *testing.T) {
	tests := []struct {
		manwon int
		want   string
	}{
		{0, "정보 없음"},
		{-1, "정보 없음"},
		{9500, "9,500만원"},
		{500, "500만원"},
		{84500, "8억 4,500만원"},
		{100000, "10억원"},
		{123456, "12억 3,456만원"},
	}
	for _, tc := range tests {
		if got := formatPrice(tc.manwon); got != tc.want {
			t.Errorf("formatPrice(%d) = %q, want %q", tc.manwon, got, tc.want)
		}
	}
}

func TestFormatBody(t *testing.T) {
	l := api.AptListing{HouseName: "테스트단지", SupplyAddress: "경기도 수원시 영통구 ..."}
	models := []api.AptModel{{
		HouseType:       "084.9500A",
		SupplyArea:      84.95,
		GeneralUnits:    120,
		SpecialUnits:    60,
		NewlywedUnits:   30,
		FirstLifeUnits:  20,
		MultiChildUnits: 0,
		TopPrice:        84500,
	}}

	t.Run("주택형 정보 포함", func(t *testing.T) {
		got := formatBody(l, models, nil)
		for _, want := range []string{
			"084.9500A", "84.95㎡", "일반공급 120세대 / 특별공급 60세대",
			"신혼부부 30, 생애최초 20", "8억 4,500만원",
		} {
			if !strings.Contains(got, want) {
				t.Errorf("본문에 %q가 없습니다:\n%s", want, got)
			}
		}
		if strings.Contains(got, "다자녀") {
			t.Errorf("0세대 유형은 생략되어야 합니다:\n%s", got)
		}
	})

	t.Run("조회 실패 시 안내 문구", func(t *testing.T) {
		got := formatBody(l, nil, errors.New("boom"))
		if !strings.Contains(got, "불러오지 못했습니다: boom") {
			t.Errorf("실패 안내가 없습니다:\n%s", got)
		}
	})

	t.Run("주택형이 없으면 안내 문구", func(t *testing.T) {
		got := formatBody(l, nil, nil)
		if !strings.Contains(got, "조회된 주택형 정보가 없습니다") {
			t.Errorf("빈 결과 안내가 없습니다:\n%s", got)
		}
	})
}
