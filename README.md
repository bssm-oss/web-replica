# Siteforge

Siteforge는 공개적으로 접근 가능한 웹 페이지를 분석해서, **원본을 그대로 복제하지 않고** 디자인 방향과 UX 흐름을 참고한 새 프론트엔드 프로젝트 생성을 돕는 Go 기반 CLI입니다.

## 이게 무엇인가요?

- 웹사이트 URL을 입력하면 HTML 구조, 제목/설명, 섹션, 네비게이션, 폼, 이미지 대체 텍스트를 분석합니다.
- Chrome/Chromium 기반 스크린샷과 computed style 샘플을 수집해 디자인 토큰 후보를 정리합니다.
- `design-spec.json` 과 `brief.md` 를 생성합니다.
- 공식 Codex CLI의 `codex exec` 를 안전하게 호출해서 새 프론트엔드 프로젝트 생성을 시도합니다.
- 생성된 프로젝트에 대해 `npm install` 과 `npm run build` 를 실행하고, 실패하면 Codex 수리 프롬프트를 한 번 더 보냅니다.

기본 모드는 **inspired reimplementation** 입니다. 로고, 브랜드명, 장문 본문, 추적 스크립트, 보호된 이미지/코드 복제를 목표로 하지 않습니다.

## 요구 사항

- Go 1.26+
- Node.js
- npm
- Git
- Chrome 또는 Chromium
- 공식 Codex CLI

## 가장 쉬운 설치 방법

```bash
git clone https://github.com/bssm-oss/web-replica.git
cd web-replica
go mod tidy
go install ./cmd/siteforge
```

Codex CLI가 없다면 먼저 설치하세요.

```bash
npm i -g @openai/codex
codex
```

위 `codex` 명령으로 공식 로그인 절차를 진행하면 됩니다. 이 프로젝트는 `~/.codex/auth.json` 을 읽지 않으며, 토큰을 직접 처리하지 않습니다.

## 빠른 시작

```bash
go run ./cmd/siteforge doctor
go run ./cmd/siteforge analyze https://example.com --out ./analysis
go run ./cmd/siteforge build https://example.com --out ./generated-site --stack vite-react-tailwind
go run ./cmd/siteforge preview ./generated-site
```

## 주요 명령

### `siteforge doctor`

로컬 개발 환경을 점검합니다.

확인 항목:

- Go
- Git
- Node.js
- npm
- Codex CLI (`codex --version`)
- Chrome/Chromium

### `siteforge analyze <url>`

웹사이트를 분석하고 다음 산출물을 만듭니다.

```text
<out>/
  .siteforge/
    latest.txt
    runs/
      <timestamp>/
        screenshots/
          desktop.png
          desktop-above-the-fold.png
          tablet.png
          tablet-above-the-fold.png
          mobile.png
          mobile-above-the-fold.png
        design-spec.json
        brief.md
        raw-outline.json
```

### `siteforge generate --spec ...`

이미 생성된 `design-spec.json` 기반으로 Codex CLI를 호출합니다.

### `siteforge build <url>`

분석 → 프롬프트 생성 → Codex 실행 → `npm install` → `npm run build` → 필요 시 수리 프롬프트까지 한 번에 수행합니다.

### `siteforge preview <path>`

- `package.json` 이 있으면 `npm run preview` 또는 `npm run dev` 실행을 시도합니다.
- 실패하면 정적 산출물(`dist/`) 직접 확인 방법을 안내합니다.

## 지원 스택

- `vite-react-tailwind` ✅ MVP 구현 완료
- `next-tailwind` ⏳ future placeholder
- `static-html-css` ⏳ future placeholder

## 보안 / 권한 정책

- `http`, `https` 외 스킴은 차단합니다.
- `localhost`, 사설 IP, loopback, link-local 대역은 기본 차단합니다.
- 리다이렉트 대상도 다시 검증합니다.
- HTML 원본 전체 저장을 기본 동작으로 두지 않습니다.
- 추적/광고/analytics/pixel 계열 asset 은 차단합니다.
- `--allow-owned-assets` 플래그가 없으면 원칙적으로 placeholder 를 사용합니다.
- Codex 인증 파일을 읽거나 출력하지 않습니다.

## 테스트 / 검증

```bash
go test ./...
go vet ./...
```

브라우저/네트워크가 준비된 환경에서는 아래 명령으로 수동 검증을 진행할 수 있습니다.

```bash
go run ./cmd/siteforge analyze https://example.com --out ./analysis
go run ./cmd/siteforge build https://example.com --out ./generated-site
```

## 폴더 구조

```text
cmd/siteforge/           CLI entrypoint
internal/cli/            Cobra commands
internal/analyzer/       URL 검증, fetch, DOM 분석, asset 필터링
internal/browser/        chromedp 스크린샷/스타일 추출
internal/spec/           design-spec / brief 생성
internal/codex/          Codex CLI 프롬프트/명령/로그 처리
internal/generator/      스택 검증 및 generate 오케스트레이션
internal/validator/      build 및 repair loop
internal/preview/        doctor / preview 실행
internal/fsutil/         안전한 경로/파일 쓰기
internal/logging/        verbose 로그와 secret masking
prompts/                 Codex 템플릿
docs/                    변경 기록과 운영 문서
testdata/                HTML fixture
```

## 알려진 제한 사항

- 브라우저 분석은 현재 chromedp 기반 MVP입니다.
- 디자인 토큰 추출은 computed style 샘플 기반의 요약이며, 완전한 디자인 시스템 역추출은 아닙니다.
- 생성 품질은 Codex 결과와 대상 사이트 구조에 영향을 받습니다.
- `preview` 는 프레임워크별 고급 런타임 관리 대신 MVP 수준 동작을 제공합니다.
- 현재 설치된 일부 Codex CLI 버전의 `codex exec` 는 approval override를 무시하고 `never` 로 동작할 수 있습니다. Siteforge는 실행 중 이 상황을 감지해 경고를 출력합니다.

## 향후 계획

- `next-tailwind` / `static-html-css` 실제 구현
- asset 다운로드 정책 세분화
- 시각적 diff 기반 검증 고도화
- 다중 페이지 분석
- 로컬 허용 대상(`--allow-local`) 명시 옵션 추가
