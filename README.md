# webreplica

URL 하나로 웹사이트를 고품질 Vite + React + TypeScript + Tailwind 프로젝트로 복제하는 CLI.

```bash
webreplica https://toss.im
```

---

## 설치

### curl (macOS / Linux — 가장 빠름)

```bash
curl -fsSL https://raw.githubusercontent.com/bssm-oss/web-replica/main/scripts/install-webreplica.sh | sh
```

### go install

```bash
go install github.com/bssm-oss/web-replica/cmd/webreplica@latest
```

### GitHub Releases에서 바이너리 직접 다운로드

[Releases 페이지](https://github.com/bssm-oss/web-replica/releases/latest)에서 OS에 맞는 파일을 받은 후:

```bash
# macOS / Linux
tar -xzf webreplica_*_darwin_arm64.tar.gz
./install.sh

# Windows (PowerShell)
Expand-Archive webreplica_*_windows_amd64.zip
.\install.ps1
```

---

## 사용법

```bash
# 기본 실행 — 고품질 클론 자동 생성
webreplica https://example.com

# 출력 폴더 지정
webreplica https://example.com --out ./my-project
```

실행하면 자동으로:

1. URL 안전성 검사
2. HTML 구조 · 섹션 · 디자인 토큰 분석
3. desktop / tablet / mobile 스크린샷 촬영
4. 사이트 이미지 · 폰트 다운로드 (같은 도메인/CDN 에셋)
5. Codex CLI로 Vite + React + TypeScript + Tailwind 프로젝트 생성
6. `npm install && npm run build` 자동 실행
7. `./generated-site` (또는 `--out` 지정 경로)에 결과 저장

```bash
# 결과 확인
cd generated-site && npm run preview
```

---

## 요구 사항

| 도구 | 설치 방법 |
|------|-----------|
| Go 1.22+ | https://go.dev/dl |
| Node.js + npm | https://nodejs.org |
| Chrome / Chromium | 대부분의 시스템에 기본 설치됨 |
| Codex CLI | `npm i -g @openai/codex` |

### Codex 로그인

```bash
webreplica login
```

또는 직접:

```bash
npm i -g @openai/codex
codex login
```

---

## 고급 옵션

```bash
# 충실도 모드 (기본값: high)
webreplica https://example.com --fidelity standard

# 출력 폴더 지정
webreplica https://example.com --out /path/to/output

# 상세 로그
webreplica https://example.com --verbose

# 분석만 실행 (생성 없음)
siteforge analyze https://example.com --out ./analysis

# 분석 결과로 생성만 실행
siteforge generate --spec ./analysis/.siteforge/runs/<timestamp>/design-spec.json --out ./generated
```

---

## 동작 방식

```
webreplica https://example.com
      │
      ├─ 1. URL 검증 + HTML fetch (chromedp)
      ├─ 2. 섹션 · 색상 · 폰트 · 이미지 분석
      ├─ 3. 동일 도메인 에셋 다운로드 → public/assets/
      ├─ 4. design-spec.json + brief.md 생성
      ├─ 5. Codex CLI → Vite + React + TS + Tailwind 프로젝트 생성
      ├─ 6. npm install && npm run build
      └─ 7. 빌드 실패 시 → Codex 수리 프롬프트 재실행
```

---

## 보안 정책

- `localhost`, 사설 IP, loopback 대역 차단
- 추적 · 광고 · analytics 에셋 차단 (`googletagmanager`, `facebook`, `hotjar` 등)
- 원본 코드 · 추적 스크립트 · 장문 콘텐츠 복사 안 함
- Codex 인증 파일 (`~/.codex/auth.json`) 읽거나 출력하지 않음

---

## 개발 / 기여

```bash
git clone https://github.com/bssm-oss/web-replica.git
cd web-replica
go mod tidy
go test ./...
go run ./cmd/webreplica https://example.com
```

```
cmd/webreplica/      CLI 진입점
internal/cli/        Cobra 커맨드
internal/analyzer/   URL 검증, fetch, DOM 분석, 에셋 필터링
internal/browser/    chromedp 스크린샷 · 스타일 추출
internal/spec/       design-spec / brief 생성
internal/codex/      Codex CLI 프롬프트 · 실행 · 로그
internal/generator/  스택 검증 및 생성 오케스트레이션
internal/validator/  빌드 및 수리 루프
prompts/             Codex 프롬프트 템플릿
```

---

## 라이선스

MIT
