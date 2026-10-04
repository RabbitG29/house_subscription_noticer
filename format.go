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
