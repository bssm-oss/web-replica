# 2026-04-22 siteforge MVP

## 배경

저장소가 비어 있었고, 안전한 웹 분석 + Codex 기반 프론트엔드 생성 CLI MVP가 요구되었다.

## 목표

- Go + Cobra 기반 `siteforge` CLI 구현
- analyze / generate / build / preview / doctor 명령 추가
- SSRF 방어, secret masking, placeholder 기본 정책 적용

## 변경 내용

- CLI 엔트리포인트와 전역 플래그 추가
- URL 검증, HTML 분석, chromedp 스크린샷/스타일 추출 구현
- design-spec / brief / raw-outline 산출물 생성
- Codex CLI 안전 실행 및 build/repair loop 추가
- 빈 컬렉션이 `null` 대신 `[]` 로 직렬화되도록 스펙 출력 정리
- owned asset 다운로드에 확장자 allowlist, redirect/same-origin 검사, host 검증 추가
- build 성공 후에도 blank page / overflow / body text 누락을 visual validation 실패로 간주하도록 repair loop 강화
- `npm ci` 실패 시 lock drift 복구를 위해 `npm install` fallback 추가
- Codex live output 스트리밍에도 secret redaction 적용
- non-interactive `codex exec` 가 로컬 MCP/플러그인 설정에 끌려가지 않도록 ephemeral + feature disable + empty `mcp_servers` override 적용
- non-interactive generator 프롬프트에 “즉시 scaffold 시작, 외부 탐색 금지” 제약 추가
- `webreplica <url>` 단축 엔트리포인트 추가. 기본 출력은 `./generated-site` 이며 전체 build 파이프라인을 바로 실행한다.
- README, AGENTS, CI, 테스트 추가

## 설계 이유

- 빈 저장소이므로 패키지 책임을 명확히 나눈 그린필드 구조를 채택했다.
- shell 문자열 조합 대신 `exec.CommandContext` arg slice를 사용해 명령 인젝션 위험을 줄였다.
- 로컬/사설망 기본 차단으로 SSRF 위험을 낮췄다.

## 영향 범위

- 저장소 전체 신규 구성

## 검증 방법

- `go test ./...`
- `go vet ./...`
- `go run ./cmd/siteforge doctor`
- 브라우저/네트워크 가능 시 `siteforge analyze` / `siteforge build`
- 수동으로 `siteforge analyze https://example.com --out <tmp>` 실행 후 `design-spec.json`, `brief.md`, 스크린샷 산출물 확인

## 남아 있는 한계

- 브라우저 기반 검증은 로컬 Chrome/Chromium 설치에 의존한다.
- Codex 품질은 모델 응답에 따라 달라질 수 있다.
- future stack은 placeholder 수준이다.
- 일부 Codex CLI 버전은 non-interactive `exec` 에서 approval override를 반영하지 않을 수 있으며, 이 경우 Siteforge는 경고만 출력할 수 있다.
- Codex 자체 stdout 은 작업 요약과 diff를 길게 출력할 수 있으며, Siteforge는 현재 이를 마스킹만 하고 축약하지는 않는다.

## 후속 과제

- Next.js 스택 지원
- 시각 diff 검증 정교화
- 멀티 페이지 분석 확장
