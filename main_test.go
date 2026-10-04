package main

import (
	"slices"
	"testing"
	"time"

	"apply-alert-bot/internal/api"
)

func TestEnvOr(t *testing.T) {
	t.Setenv("TEST_ENV_OR_KEY", "값있음")

	if got := envOr("TEST_ENV_OR_KEY", "기본값"); got != "값있음" {
		t.Errorf("envOr(설정된 키) = %q, want %q", got, "값있음")
	}

	if got := envOr("TEST_ENV_OR_KEY_NOT_SET", "기본값"); got != "기본값" {
		t.Errorf("envOr(없는 키) = %q, want %q", got, "기본값")
	}
}

func TestSplitAndTrim(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{"빈 문자열은 nil", "", nil},
		{"단일 값", "수원시", []string{"수원시"}},
		{"쉼표로 분리하고 공백은 트림", " 수원시 , 용인시 ", []string{"수원시", "용인시"}},
		{"빈 조각(연속 쉼표)은 버림", "수원시,,용인시", []string{"수원시", "용인시"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := splitAndTrim(tc.input)
			if !slices.Equal(got, tc.want) {
				t.Errorf("splitAndTrim(%q) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}

func TestParseDurationOr(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		fallback time.Duration
		want     time.Duration
	}{
		{"빈 문자열이면 기본값 사용", "", 6 * time.Hour, 6 * time.Hour},
		{"정상적인 값은 그대로 파싱", "30m", time.Hour, 30 * time.Minute},
		{"파싱 안 되는 값이면 기본값으로 폴백", "이상한값", time.Hour, time.Hour},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := parseDurationOr(tc.input, tc.fallback)
			if got != tc.want {
				t.Errorf("parseDurationOr(%q, %v) = %v, want %v", tc.input, tc.fallback, got, tc.want)
			}
		})
	}
}

func TestMatchesRegion(t *testing.T) {
	tests := []struct {
		name    string
		listing api.AptListing
		filters []string
		want    bool
	}{
		{
			name:    "필터가 없으면 전부 통과",
			listing: api.AptListing{SupplyAddress: "경기도 수원시 영통구 ..."},
			filters: nil,
			want:    true,
		},
		{
			name:    "SupplyAddress에 필터 문자열이 포함되면 통과",
			listing: api.AptListing{SupplyAddress: "경기도 수원시 영통구 ..."},
			filters: []string{"영통구"},
			want:    true,
		},
		{
			name:    "SupplyAddress가 비어있으면 RegionName으로 대체",
			listing: api.AptListing{RegionName: "경기"},
			filters: []string{"경기"},
			want:    true,
		},
		{
			name:    "필터와 안 맞으면 탈락",
			listing: api.AptListing{SupplyAddress: "울산광역시 남구 ..."},
			filters: []string{"수원시", "용인시"},
			want:    false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := matchesRegion(tc.listing, tc.filters)
			if got != tc.want {
				t.Errorf("matchesRegion() = %v, want %v", got, tc.want)
			}
		})
	}
}

// TestMatchesDate는 "오늘" 날짜를 기준으로 상대적인 날짜 문자열을 만들어
// 테스트합니다. matchesDate가 내부적으로 time.Now()를 직접 호출해서
// "오늘 날짜"를 하드코딩할 수 없기 때문입니다(하드코딩하면 테스트를 실행하는
// 날짜가 바뀔 때마다 테스트가 깨지는 flaky test가 됩니다).
func TestMatchesDate(t *testing.T) {
	loc := time.UTC
	now := time.Now().In(loc)
	today := now.Format("2006-01-02")
	yesterday := now.AddDate(0, 0, -1).Format("2006-01-02")
	tomorrow := now.AddDate(0, 0, 1).Format("2006-01-02")

	tests := []struct {
		name       string
		receiptEnd string
		want       bool
	}{
		{"마감일이 미래면 아직 열려있음", tomorrow, true},
		{"마감일이 오늘이면 당일까지는 열려있음", today, true},
		{"마감일이 과거면 이미 마감됨", yesterday, false},
		{"형식이 이상하면 안전하게 열려있다고 취급(fail-open)", "이상한값", true},
		{"빈 문자열도 안전하게 열려있다고 취급(fail-open)", "", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			listing := api.AptListing{ReceiptEnd: tc.receiptEnd}
			got := matchesDate(listing, loc)
			if got != tc.want {
				t.Errorf("matchesDate(ReceiptEnd=%q) = %v, want %v", tc.receiptEnd, got, tc.want)
			}
		})
	}
}

func TestCompetitionDue(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Seoul")
	at := func(s string) time.Time {
		tm, err := time.ParseInLocation("2006-01-02 15:04", s, loc)
		if err != nil {
			t.Fatal(err)
		}
		return tm
	}
	l := api.AptListing{ReceiptEnd: "2026-10-14"}

	tests := []struct {
		name string
		now  time.Time
		want bool
	}{
		{"종료 당일은 아직", at("2026-10-14 23:59"), false},
		{"종료 다음 날 0시부터", at("2026-10-15 00:00"), true},
		{"창 마지막 날", at("2026-10-28 23:59"), true},
		{"창 밖(15일째)", at("2026-10-29 00:00"), false},
		{"접수 중", at("2026-10-10 12:00"), false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := competitionDue(l, tc.now, loc); got != tc.want {
				t.Errorf("competitionDue(now=%v) = %v, want %v", tc.now, got, tc.want)
			}
		})
	}

	if competitionDue(api.AptListing{ReceiptEnd: ""}, at("2026-10-15 00:00"), loc) {
		t.Error("종료일을 파싱할 수 없으면 대상에서 빠져야 함")
	}
}
