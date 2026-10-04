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
	competitionDBPath := envOr("COMPETITION_SEEN_DB_PATH", "./competition-seen.json")
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
	competitionSeen, err := store.Load(competitionDBPath)
	if err != nil {
		log.Fatalf("경쟁률 상태 파일 불러오기 실패: %v", err)
	}

	if runOnceOnly {
		log.Printf("청약 알림봇 단발 실행(RUN_ONCE) — 지역 필터: %v", regionFilter)
		runOnce(client, mailer, seen, competitionSeen, regionFilter, loc)
		return
	}

	log.Printf("청약 알림봇 시작 — 지역 필터: %v, 폴링 주기: %s", regionFilter, pollInterval)

	runOnce(client, mailer, seen, competitionSeen, regionFilter, loc)

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()
	for range ticker.C {
		runOnce(client, mailer, seen, competitionSeen, regionFilter, loc)
	}
}

func runOnce(client *api.Client, mailer *notify.Email, seen, competitionSeen *store.Store, regionFilter []string, loc *time.Location) {
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
		// 주택형별 정보는 부가 정보라서, 조회에 실패해도 공고 알림 자체는 보냅니다.
		// (실패하면 seen에 기록되므로 다음 폴링에서 재시도되지 않는다는 점은 감수)
		models, err := client.FetchModelsByAnnouncement(l.AnnouncementNo)
		if err != nil {
			log.Printf("주택형별 정보 조회 실패 (%s) — 기본 정보만 발송합니다: %v", l.HouseName, err)
		}
		body := buildBody(l, models)
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

	notifyCompetitions(client, mailer, competitionSeen, listings, regionFilter, loc)
}

// competitionWindowDays는 접수 종료 후 경쟁률 알림 대상으로 보는 기간입니다.
// 이 기간 밖의 공고는 후보에서 빠지므로, 처음 실행해도 과거 공고의 경쟁률이
// 한꺼번에 발송되지 않고 경쟁률 상태 파일도 무한정 커지지 않습니다.
const competitionWindowDays = 14

// notifyCompetitions는 접수가 끝난 관심 지역 공고의 일반공급 경쟁률과
// 특별공급 신청현황을 공고당 메일 한 통으로 보냅니다. 경쟁률이 아직 집계되지
// 않았으면(빈 응답) 기록하지 않고 다음 폴링에서 다시 시도합니다.
func notifyCompetitions(client *api.Client, mailer *notify.Email, competitionSeen *store.Store, listings []api.AptListing, regionFilter []string, loc *time.Location) {
	now := time.Now().In(loc)
	sent := 0
	for _, l := range listings {
		if !matchesRegion(l, regionFilter) || !competitionDue(l, now, loc) {
			continue
		}
		if competitionSeen.Has(l.AnnouncementNo) {
			continue
		}

		comps, err := client.FetchCompetitionByAnnouncement(l.AnnouncementNo)
		if err != nil {
			log.Printf("경쟁률 조회 실패 (%s): %v — 다음 폴링에서 재시도합니다", l.HouseName, err)
			continue
		}
		if len(comps) == 0 {
			log.Printf("경쟁률 미집계 (%s) — 다음 폴링에서 재시도합니다", l.HouseName)
			continue
		}
		// 특별공급 신청현황은 부가 정보라서, 조회에 실패해도 경쟁률 알림 자체는 보냅니다.
		specials, err := client.FetchSpecialStatusByAnnouncement(l.AnnouncementNo)
		if err != nil {
			log.Printf("특별공급 신청현황 조회 실패 (%s) — 일반공급 경쟁률만 발송합니다: %v", l.HouseName, err)
		}

		subject := fmt.Sprintf("📊 청약 경쟁률: %s", l.HouseName)
		if err := mailer.Send(subject, buildCompetitionBody(l, comps, specials)); err != nil {
			log.Printf("경쟁률 알림 전송 실패 (%s): %v — 다음 폴링에서 재시도합니다", l.HouseName, err)
			continue
		}

		competitionSeen.MarkSeen(l.AnnouncementNo)
		sent++
	}

	if sent > 0 {
		if err := competitionSeen.Save(); err != nil {
			log.Printf("경쟁률 상태 저장 실패: %v", err)
		}
	}
	log.Printf("경쟁률 알림 — %d건 발송", sent)
}

// competitionDue는 접수 종료일 다음 날부터 competitionWindowDays일째까지인지
// 확인합니다. 종료 당일에는 집계가 덜 끝났을 수 있어 하루 지난 뒤에 보냅니다.
// 종료일을 파싱할 수 없으면 언제까지가 창인지 알 수 없으므로 대상에서 뺍니다.
func competitionDue(l api.AptListing, now time.Time, loc *time.Location) bool {
	end, err := time.ParseInLocation("2006-01-02", l.ReceiptEnd, loc)
	if err != nil {
		return false
	}
	from := end.AddDate(0, 0, 1)
	until := end.AddDate(0, 0, 1+competitionWindowDays)
	return !now.Before(from) && now.Before(until)
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
