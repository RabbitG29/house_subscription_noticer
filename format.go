package main

import (
	"fmt"
	"strconv"
	"strings"

	"apply-alert-bot/internal/api"
)

// buildBody는 공고 한 건의 메일 본문을 만듭니다. models가 비어 있으면
// (주택형 조회 실패 포함) 주택형별 섹션 없이 기본 정보만 담깁니다.
func buildBody(l api.AptListing, models []api.AptModel) string {
	body := fmt.Sprintf(
		"공급위치: %s\n주택유형: %s (%s)\n총 공급세대수: %d세대\n모집공고일: %s\n접수기간: %s ~ %s\n당첨자발표일: %s\n%s",
		l.SupplyAddress, l.HouseTypeName, l.HouseDetailType, l.TotalSupplyUnits,
		l.NoticeDate, l.ReceiptStart, l.ReceiptEnd, l.WinnerAnnounceDate, l.HomepageURL,
	)
	if len(models) > 0 {
		body += "\n\n" + formatModels(models)
	}
	return body
}

// formatModels는 주택형별 분양가와 일반/특별공급 세대수를 한 블록씩 표시합니다.
func formatModels(models []api.AptModel) string {
	var b strings.Builder
	b.WriteString("[주택형별 공급 정보]")
	for _, m := range models {
		fmt.Fprintf(&b, "\n\n■ %s (공급면적 %s㎡)", formatHouseType(m.HouseType), m.SupplyArea)
		fmt.Fprintf(&b, "\n  분양가(최고가): %s", formatPrice(m.TopPrice))
		fmt.Fprintf(&b, "\n  일반공급: %d세대", m.GeneralUnits)
		fmt.Fprintf(&b, "\n  특별공급: %d세대", m.SpecialUnits)
		if detail := specialBreakdown(m); detail != "" {
			fmt.Fprintf(&b, " (%s)", detail)
		}
	}
	return b.String()
}

// specialBreakdown은 특별공급 유형 중 세대수가 0보다 큰 것만 "신혼부부 5, 생애최초 3"
// 형태로 이어 붙입니다. 0인 유형까지 전부 나열하면 메일이 지저분해집니다.
func specialBreakdown(m api.AptModel) string {
	parts := []struct {
		name  string
		units int
	}{
		{"다자녀", m.MultiChildUnits},
		{"신혼부부", m.NewlywedUnits},
		{"생애최초", m.FirstLifeUnits},
		{"노부모부양", m.OldParentsUnits},
		{"기관추천", m.InstitutionUnits},
		{"청년", m.YoungUnits},
		{"신생아", m.NewbornUnits},
		{"이전기관", m.TransferAgencyUnit},
		{"기타", m.EtcUnits},
	}
	var out []string
	for _, p := range parts {
		if p.units > 0 {
			out = append(out, fmt.Sprintf("%s %d", p.name, p.units))
		}
	}
	return strings.Join(out, ", ")
}

// formatPrice는 만원 단위 문자열("36707")을 "3억 6,707만원"으로 바꿉니다.
// 숫자로 해석할 수 없으면 원문을 그대로 보여줘서 정보가 사라지지 않게 합니다.
func formatPrice(manwon string) string {
	n, err := strconv.Atoi(strings.TrimSpace(manwon))
	if err != nil || n <= 0 {
		if manwon == "" {
			return "정보 없음"
		}
		return manwon
	}
	eok, rest := n/10000, n%10000
	switch {
	case eok == 0:
		return withComma(rest) + "만원"
	case rest == 0:
		return fmt.Sprintf("%d억원", eok)
	default:
		return fmt.Sprintf("%d억 %s만원", eok, withComma(rest))
	}
}

func withComma(n int) string {
	s := strconv.Itoa(n)
	if len(s) <= 3 {
		return s
	}
	return s[:len(s)-3] + "," + s[len(s)-3:]
}

// formatHouseType은 "084.9500A"를 "84A"로 정리합니다(앞자리 0과 소수 0 제거,
// 뒤의 타입 문자는 유지). 패턴과 안 맞으면 원문을 그대로 돌려줍니다.
func formatHouseType(raw string) string {
	i := 0
	for i < len(raw) && (raw[i] == '.' || (raw[i] >= '0' && raw[i] <= '9')) {
		i++
	}
	f, err := strconv.ParseFloat(raw[:i], 64)
	if err != nil {
		return raw
	}
	return strconv.FormatFloat(f, 'f', -1, 64) + raw[i:]
}

// buildCompetitionBody는 접수가 끝난 공고의 경쟁률 메일 본문을 만듭니다.
// 특별공급 신청현황은 부가 정보라 specials가 비어 있으면 그 섹션만 생략합니다.
func buildCompetitionBody(l api.AptListing, comps []api.AptCompetition, specials []api.AptSpecialStatus) string {
	body := fmt.Sprintf("공급위치: %s\n접수기간: %s ~ %s\n당첨자발표일: %s\n%s\n\n%s",
		l.SupplyAddress, l.ReceiptStart, l.ReceiptEnd, l.WinnerAnnounceDate, l.HomepageURL,
		formatCompetition(comps))
	if len(specials) > 0 {
		body += "\n\n" + formatSpecialStatus(specials)
	}
	return body
}

// formatCompetition은 일반공급 경쟁률을 주택형별 블록으로 묶습니다. 접수가 0건인
// 행(예: 1순위에서 마감돼 접수가 없는 2순위)은 의미가 없어 뺍니다.
// 주택형은 API가 내려준 순서를 그대로 유지합니다.
func formatCompetition(comps []api.AptCompetition) string {
	var order []string
	byType := make(map[string][]api.AptCompetition)
	for _, c := range comps {
		if _, ok := byType[c.HouseType]; !ok {
			order = append(order, c.HouseType)
		}
		byType[c.HouseType] = append(byType[c.HouseType], c)
	}

	var b strings.Builder
	b.WriteString("[일반공급 경쟁률]")
	for _, ht := range order {
		rows := byType[ht]
		fmt.Fprintf(&b, "\n\n■ %s (일반공급 %d세대)", formatHouseType(ht), rows[0].SupplyUnits)
		printed := 0
		for _, c := range rows {
			if strings.TrimSpace(c.RequestCount) == "0" {
				continue
			}
			fmt.Fprintf(&b, "\n  %d순위 %s: 접수 %s건 · %s",
				c.Rank, c.ResideName, c.RequestCount, formatRate(c.Rate))
			printed++
		}
		if printed == 0 {
			b.WriteString("\n  접수 없음")
		}
	}
	return b.String()
}

// formatRate는 경쟁률 원문을 읽기 쉽게 바꿉니다. "7.28" → "7.28:1",
// "(△1)" → "1세대 미달", "-"(상위 지역에서 마감되어 산정 안 됨)는 그대로 둡니다.
func formatRate(rate string) string {
	rate = strings.TrimSpace(rate)
	switch {
	case rate == "" || rate == "-":
		return "경쟁률 미산정"
	case strings.HasPrefix(rate, "(△") && strings.HasSuffix(rate, ")"):
		return strings.TrimSuffix(strings.TrimPrefix(rate, "(△"), ")") + "세대 미달"
	default:
		if _, err := strconv.ParseFloat(rate, 64); err == nil {
			return rate + ":1"
		}
		return rate
	}
}

// specialRow는 특별공급 유형 하나의 공급세대수와 거주지역별 접수건수입니다.
type specialRow struct {
	name                 string
	units                int
	local, state, others int
}

func specialRows(s api.AptSpecialStatus) []specialRow {
	return []specialRow{
		{"다자녀", s.MultiChildUnits, s.MultiChildLocal, s.MultiChildState, s.MultiChildOther},
		{"신혼부부", s.NewlywedUnits, s.NewlywedLocal, s.NewlywedState, s.NewlywedOther},
		{"생애최초", s.FirstLifeUnits, s.FirstLifeLocal, s.FirstLifeState, s.FirstLifeOther},
		{"노부모부양", s.OldParentsUnits, s.OldParentsLocal, s.OldParentsState, s.OldParentsOther},
		{"청년", s.YoungUnits, s.YoungLocal, s.YoungState, s.YoungOther},
		{"신생아", s.NewbornUnits, s.NewbornLocal, s.NewbornState, s.NewbornOther},
	}
}

// formatSpecialStatus는 특별공급 신청현황을 주택형별로, 세대수나 접수가 있는
// 유형만 표시합니다. 접수 합계를 공급세대수로 나눈 값은 API가 주는 공식
// 경쟁률이 아니라 대략적인 참고치입니다(수치가 소수 첫째 자리).
func formatSpecialStatus(specials []api.AptSpecialStatus) string {
	var b strings.Builder
	b.WriteString("[특별공급 신청현황]")
	for _, s := range specials {
		fmt.Fprintf(&b, "\n\n■ %s (특별공급 %d세대)", formatHouseType(s.HouseType), s.TotalUnits)
		lines := 0
		for _, r := range specialRows(s) {
			total := r.local + r.state + r.others
			if r.units == 0 && total == 0 {
				continue
			}
			fmt.Fprintf(&b, "\n  %s %d세대: 접수 %d건 (해당지역 %d, 해당 시·도 %d, 기타 %d)%s",
				r.name, r.units, total, r.local, r.state, r.others, approxRate(total, r.units))
			lines++
		}
		if s.InstitutionUnits > 0 || s.InstitutionDecided+s.InstitutionPrepared > 0 {
			fmt.Fprintf(&b, "\n  기관추천 %d세대: 결정 %d건, 준비 %d건",
				s.InstitutionUnits, s.InstitutionDecided, s.InstitutionPrepared)
			lines++
		}
		if s.TransferUnits > 0 || s.TransferCount > 0 {
			fmt.Fprintf(&b, "\n  이전기관 %d세대: 접수 %d건%s",
				s.TransferUnits, s.TransferCount, approxRate(s.TransferCount, s.TransferUnits))
			lines++
		}
		if lines == 0 {
			b.WriteString("\n  접수 없음")
		}
	}
	return b.String()
}

// approxRate는 접수건수/공급세대수를 " · 약 3.5:1"로 만들고, 세대수가 0이면 빈 문자열입니다.
func approxRate(count, units int) string {
	if units <= 0 {
		return ""
	}
	return fmt.Sprintf(" · 약 %.1f:1", float64(count)/float64(units))
}
