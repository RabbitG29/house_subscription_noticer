package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"apply-alert-bot/internal/api"
	"apply-alert-bot/internal/notify"
	"apply-alert-bot/internal/store"
)

func main() {
	serviceKey := requireEnv("DATA_GO_KR_SERVICE_KEY")
	smtpUsername := requireEnv("SMTP_USERNAME")
	smtpPassword := requireEnv("SMTP_PASSWORD")
	emailTo := requireEnv("EMAIL_TO")

	regionFilter := splitAndTrim(os.Getenv("REGION_FILTER")) // e.g. "경기,수원,용인"
	dbPath := envOr("SEEN_DB_PATH", "./seen.json")
	pollInterval := parseDurationOr(os.Getenv("POLL_INTERVAL"), 6*time.Hour)
	runOnceOnly := os.Getenv("RUN_ONCE") == "true"

	loc, locErr := time.LoadLocation("Asia/Seoul")
	if locErr != nil {
		log.Printf("타임존 로드 실패, UTC로 대체합니다: %v", locErr)
		loc = time.UTC
	}

	client := api.NewClient(serviceKey)
	mailer := &notify.Email{
		Host:     envOr("SMTP_HOST", "smtp.naver.com"),
		Port:     envOr("SMTP_PORT", "587"),
		Username: smtpUsername,
		Password: smtpPassword,
		From:     envOr("EMAIL_FROM", smtpUsername),
		To:       emailTo,
	}

	seen, err := store.Load(dbPath)
	if err != nil {
		log.Fatalf("상태 파일 불러오기 실패: %v", err)
	}

	if runOnceOnly {
		log.Printf("청약 알림봇 단발 실행(RUN_ONCE) — 지역 필터: %v", regionFilter)
		runOnce(client, mailer, seen, regionFilter, loc)
		return
	}

	log.Printf("청약 알림봇 시작 — 지역 필터: %v, 폴링 주기: %s", regionFilter, pollInterval)

	runOnce(client, mailer, seen, regionFilter, loc)

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()
	for range ticker.C {
		runOnce(client, mailer, seen, regionFilter, loc)
	}
}

func runOnce(client *api.Client, mailer *notify.Email, seen *store.Store, regionFilter []string, loc *time.Location) {
	listings, err := client.FetchAllAptListings()
	if err != nil {
		log.Printf("API 조회 실패: %v", err)
		if notifyErr := mailer.Send("⚠️ 청약 API 조회 실패", fmt.Sprintf("%v", err)); notifyErr != nil {
			log.Printf("실패 알림 전송도 실패: %v", notifyErr)
		}
		return
	}

	newCount := 0
	for _, l := range listings {
		if !matchesRegion(l, regionFilter) {
			continue
		}
		if !matchesDate(l, loc) {
			continue
		}
		if seen.Has(l.AnnouncementNo) {
			continue
		}

		subject := fmt.Sprintf("🏠 신규 청약 공고: %s", l.HouseName)
		body := fmt.Sprintf(
			"공급위치: %s\n주택유형: %s (%s)\n총 공급세대수: %d세대\n모집공고일: %s\n접수기간: %s ~ %s\n당첨자발표일: %s\n%s",
			l.SupplyAddress, l.HouseTypeName, l.HouseDetailType, l.TotalSupplyUnits,
			l.NoticeDate, l.ReceiptStart, l.ReceiptEnd, l.WinnerAnnounceDate, l.HomepageURL,
		)
		if err := mailer.Send(subject, body); err != nil {
			log.Printf("알림 전송 실패 (%s): %v — 다음 폴링에서 재시도합니다", l.HouseName, err)
			continue
		}

		seen.MarkSeen(l.AnnouncementNo)
		newCount++
	}

	if newCount > 0 {
		if err := seen.Save(); err != nil {
			log.Printf("상태 저장 실패: %v", err)
		}
	}
	log.Printf("조회 완료 — 전체 %d건, 신규 %d건", len(listings), newCount)
}

// matchesRegion은 SupplyAddress(시/구 단위 상세주소, 예: "경기도 수원시
// 권선구 ...")에 필터 문자열이 포함되는지 확인합니다. SupplyAddress가
// 비어 있는 경우에만 RegionName(시/도 단위)으로 대체합니다.
func matchesRegion(l api.AptListing, filters []string) bool {
	if len(filters) == 0 {
		return true
	}
	haystack := l.SupplyAddress
	if haystack == "" {
		haystack = l.RegionName
	}
	for _, f := range filters {
		if strings.Contains(haystack, f) {
			return true
		}
	}
	return false
}

// matchesDate는 현재 날짜보다 청약 접수 종료일이 같거나 미래인지 확인합니다.
func matchesDate(l api.AptListing, loc *time.Location) bool {
	now := time.Now().In(loc)
	lastDate, err := time.ParseInLocation("2006-01-02", l.ReceiptEnd, loc)
	if err != nil {
		return true
	}
	if now.Before(lastDate.Add(24 * time.Hour)) {
		return true
	}

	return false
}

func requireEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("환경변수 %s가 설정되지 않았습니다.", key)
	}
	return v
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func splitAndTrim(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func parseDurationOr(s string, fallback time.Duration) time.Duration {
	if s == "" {
		return fallback
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		log.Printf("POLL_INTERVAL 파싱 실패(%s), 기본값 사용: %v", s, fallback)
		return fallback
	}
	return d
}
