# AGENTS.md

## 프로젝트 목적

Siteforge는 공개 웹페이지를 분석하고, 안전한 inspired reimplementation 프론트엔드 생성을 돕는 Go CLI다.

## 빠른 시작 명령

```bash
go mod tidy
go test ./...
go run ./cmd/siteforge doctor
```

## 설치 / 실행 / 테스트 명령

- 설치: `go install ./cmd/siteforge`
- 환경 점검: `go run ./cmd/siteforge doctor`
- 분석: `go run ./cmd/siteforge analyze https://example.com --out ./analysis`
- 전체 파이프라인: `go run ./cmd/siteforge build https://example.com --out ./generated-site`
- 테스트: `go test ./...`
- 정적 검사: `go vet ./...`

## 기본 작업 순서

1. README, AGENTS, docs 변경점 확인
2. 관련 패키지와 테스트를 함께 수정
3. 보안 규칙(SSRF, secret masking, Codex auth 비접근)을 유지
4. `gofmt`, `go test`, `go vet` 실행
5. README / docs 반영

## 완료 조건

- 요청 기능 반영
- 관련 테스트 추가 또는 갱신
- `go test ./...` 통과
- `go vet ./...` 통과
- README / docs 최신화
- 시크릿 출력 없음

## 코드 스타일 원칙

- 과도한 추상화보다 읽기 쉬운 구조를 우선
- shell 문자열 합치기 금지, `exec.CommandContext` + arg slice 사용
- 오류 메시지는 다음 행동이 보이도록 작성
- panic 대신 반환 가능한 오류 사용

## 파일 구조 원칙

- CLI: `internal/cli`
- 분석: `internal/analyzer`
- 브라우저: `internal/browser`
- Codex 통합: `internal/codex`
- 검증: `internal/validator`
- 문서: `docs/changes`

## 문서화 원칙

- 의미 있는 변경은 `docs/changes/` 에 기록
- README는 한글 유지
- 법적/보안 제약이 바뀌면 README와 docs를 함께 수정

## 테스트 원칙

- URL validation, asset filtering, prompt/command safety, HTML parsing은 회귀 테스트 필수
- 외부 서비스 의존 없이 가능한 한 fixture 기반으로 작성

## 브랜치 / 커밋 / PR 규칙

- 브랜치명 예: `feat/siteforge-mvp`
- 커밋은 기능/테스트/문서를 분리
- PR 본문에는 배경, 변경 내용, 테스트 결과, 리스크를 포함

## 민감한 경로 / 주의 경로

- `internal/codex/`: secret masking 유지
- `internal/analyzer/url.go`: SSRF 보호 유지
- `prompts/`: 불법 복제 유도 문구 추가 금지

## 작업 전 체크리스트

- 현재 저장소 상태 확인
- 관련 테스트 위치 확인
- 사용자 요청 범위 재확인

## 작업 후 체크리스트

- `gofmt` 실행
- `go test ./...`
- `go vet ./...`
- README / docs 반영

## 절대 하면 안 되는 것

- Codex auth 파일 읽기
- 토큰/쿠키/Authorization 로그 출력
- shell injection 가능 코드 추가
- localhost/private IP 분석 허용 기본값 추가
- 저작권 보호 콘텐츠 복제 기능 추가
