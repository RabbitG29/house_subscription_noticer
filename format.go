package main

import (
	"fmt"
	"strings"

	"apply-alert-bot/internal/api"
)

// formatBody builds the notification mail body. models가 비어 있으면 주택형
// 섹션 대신 안내 문구를 넣고, modelErr가 있으면 조회 실패 사실을 알립니다.
func formatBody(l api.AptListing, models []api.AptModel, modelErr error) string {
	var b strings.Builder
	fmt.Fprintf(&b,
		"공급위치: %s\n주택유형: %s (%s)\n총 공급세대수: %d세대\n모집공고일: %s\n접수기간: %s ~ %s\n당첨자발표일: %s\n%s\n",
		l.SupplyAddress, l.HouseTypeName, l.HouseDetailType, l.TotalSupplyUnits,
		l.NoticeDate, l.ReceiptStart, l.ReceiptEnd, l.WinnerAnnounceDate, l.HomepageURL,
	)

	b.WriteString("\n[주택형별 정보]\n")
	switch {
	case modelErr != nil:
		fmt.Fprintf(&b, "주택형 정보를 불러오지 못했습니다: %v\n", modelErr)
	case len(models) == 0:
		b.WriteString("조회된 주택형 정보가 없습니다.\n")
	default:
		for _, m := range models {
			b.WriteString(formatModel(m))
		}
	}
	return b.String()
}

func formatModel(m api.AptModel) string {
	var b strings.Builder
	fmt.Fprintf(&b, "\n■ %s (공급면적 %.2f㎡)\n", m.HouseType, float64(m.SupplyArea))
	fmt.Fprintf(&b, "  일반공급 %d세대 / 특별공급 %d세대\n", m.GeneralUnits.Int(), m.SpecialUnits.Int())

	if parts := m.SpecialBreakdown(); len(parts) > 0 {
		names := make([]string, 0, len(parts))
		for _, p := range parts {
			names = append(names, fmt.Sprintf("%s %d", p.Name, p.Units))
		}
		fmt.Fprintf(&b, "    - 특공 내역: %s\n", strings.Join(names, ", "))
	}
	fmt.Fprintf(&b, "  최고 분양가: %s\n", formatPrice(m.TopPrice.Int()))
	return b.String()
}

// formatPrice formats an amount given in 만원 (e.g. 84500 → "8억 4,500만원").
// 청약홈 분양가 단위가 만원이라는 가정은 실제 응답으로 확인이 필요합니다.
func formatPrice(manwon int) string {
	if manwon <= 0 {
		return "정보 없음"
	}
	eok, rest := manwon/10000, manwon%10000
	switch {
	case eok == 0:
		return withCommas(rest) + "만원"
	case rest == 0:
		return fmt.Sprintf("%d억원", eok)
	default:
		return fmt.Sprintf("%d억 %s만원", eok, withCommas(rest))
	}
}

func withCommas(n int) string {
	s := fmt.Sprintf("%d", n)
	var out []byte
	for i, c := range []byte(s) {
		if i > 0 && (len(s)-i)%3 == 0 {
			out = append(out, ',')
		}
		out = append(out, c)
	}
	return string(out)
}
