# apply-alert-bot

한국부동산원 청약홈 "APT 분양정보 상세조회" 공식 오픈API를 폴링해서,
관심 지역에 신규 공고가 뜨면 이메일(SMTP)로 알려주는 봇입니다.
외부 Go 모듈 의존성 없이 표준 라이브러리만 사용합니다 (DB 대신 JSON
파일로 중복 알림을 방지, 발송도 net/smtp로 직접 처리).

## 실행 전 확인해야 할 것

1. **필드명 검증** — `internal/api/types.go`의 struct 태그는 공개된
   문서/예제를 바탕으로 작성했고, 확정 문서로 재확인하지 못했습니다.
   승인된 서비스키로 아래 `cmd/inspect` 도구를 한 번 돌려서, 응답
   JSON의 실제 필드명이 태그와 정확히 일치하는지 (특히 "공급위치
   (상세주소)"에 해당하는 필드가 있는지) 직접 확인하세요:

   ```bash
   DATA_GO_KR_SERVICE_KEY=발급받은키 go run ./cmd/inspect
   ```

   원본 JSON을 그대로 pretty-print해서 보여줍니다. 현재는 시/도
   단위인 `SUBSCRPT_AREA_CODE_NM`만 필터링에 쓰고 있어서 "수원"/"용인"
   같은 시/구 단위까지는 못 거를 수 있습니다 — 상세주소 필드가
   확인되면 `internal/api/types.go`에 필드를 추가하고
   `main.go`의 `matchesRegion`이 그 필드를 우선 쓰도록 바꾸세요.

2. **메일 계정 준비** — 네이버 메일 기준, 메일 > 환경설정 > POP3/IMAP
   설정에서 SMTP 사용을 켜고, 2단계 인증을 쓰는 계정이면 앱 비밀번호를
   발급받아 `SMTP_PASSWORD`에 넣으세요(로그인 비밀번호 아님). Gmail 등
   다른 제공자를 쓰려면 `.env`의 `SMTP_HOST`/`SMTP_PORT`만 바꾸면
   됩니다 (`net/smtp`가 STARTTLS를 알아서 협상합니다).

3. **트래픽 한도** — 개발계정 트래픽 한도 내에서 폴링 주기를
   설정하세요 (기본값 6시간).

## 실행 방법

```bash
cp .env.example .env
# .env 값 채운 후
set -a; source .env; set +a
go run .
```

(`REGION_FILTER`처럼 공백이 든 값은 `export $(cat .env | xargs)` 방식으로는
깨질 수 있으니 위처럼 `source`를 쓰세요.)

한 번만 실행하고 종료하려면 (배포용) `RUN_ONCE=true`를 추가하세요.

## 상시 구동: GitHub Actions cron

개인 서버/라즈베리파이 없이도 돌릴 수 있고, 사내(NC소프트) 인프라를
쓰지 않는다는 제약과도 맞아서 GitHub Actions의 scheduled workflow로
구동하도록 구성했습니다 (`.github/workflows/poll.yml`, 매 6시간 실행).
중복 알림 방지용 `seen.json`은 매 실행 후 변경이 있으면 워크플로우가
직접 리포지토리에 커밋합니다.

설정 방법:

```bash
gh secret set DATA_GO_KR_SERVICE_KEY
gh secret set SMTP_USERNAME
gh secret set SMTP_PASSWORD
gh secret set EMAIL_TO
gh variable set REGION_FILTER --body "경기,수원,용인"
```

푸시 후 Actions 탭에서 "Run workflow"로 수동 실행해 먼저 한 번
확인해보세요. 폴링 주기를 바꾸려면 `poll.yml`의 `cron` 표현식을
수정하면 됩니다 (공공데이터포털 개발계정 트래픽 한도 고려).

다른 대안(로컬 상시 실행, VM/라즈베리파이 systemd)도 고려했지만,
별도 하드웨어/서버 유지보수가 필요 없고 이미 쓰는 GitHub 리포지토리
안에서 완결된다는 점에서 GitHub Actions cron을 선택했습니다. 실행
로그가 공개 Actions 탭에 남는 게 싫다면 리포지토리를 private으로
두거나 systemd 방식으로 바꾸세요.

## 다음에 손볼 만한 부분

- 상세 주소 필드 확인 후 지역 필터를 시/구 단위로 세분화
- "청약접수 경쟁률 및 특별공급 신청현황 조회 서비스" API 연동 (접수
  마감 후 결과 확인용, 엔드포인트는 `ApplyhomeInfoDetailSvc` 계열
  네이밍 규칙을 참고해 Swagger에서 확인 — 추측 금지)
- JSON 파일 → SQLite로 교체 (동시 실행이나 이력 조회가 필요해지면)
