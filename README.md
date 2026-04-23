# Siteforge

Siteforge는 공개적으로 접근 가능한 웹 페이지를 분석해서, **원본을 그대로 복제하지 않고** 디자인 방향과 UX 흐름을 참고한 새 프론트엔드 프로젝트 생성을 돕는 Go 기반 CLI입니다.

## 이게 무엇인가요?

- 웹사이트 URL을 입력하면 HTML 구조, 제목/설명, 섹션, 네비게이션, 폼, 이미지 대체 텍스트를 분석합니다.
- Chrome/Chromium 기반 스크린샷과 computed style 샘플을 수집해 디자인 토큰 후보를 정리합니다.
- `design-spec.json` 과 `brief.md` 를 생성합니다.
- 공식 Codex CLI의 `codex exec` 를 안전하게 호출해서 새 프론트엔드 프로젝트 생성을 시도합니다.
- 생성된 프로젝트에 대해 `npm install` 과 `npm run build` 를 실행하고, 실패하면 Codex 수리 프롬프트를 한 번 더 보냅니다.

기본 모드는 **inspired reimplementation** 입니다. 로고, 브랜드명, 장문 본문, 추적 스크립트, 보호된 이미지/코드 복제를 목표로 하지 않습니다.

## 한 줄 실행

처음 한 번만 설치하면, 이후에는 어느 폴더에서든 아래 한 줄만 치면 됩니다.

```bash
webreplica https://원하는사이트.com
```

`./webreplica` 처럼 현재 폴더 실행 파일 경로를 붙일 필요가 없습니다. GitHub Release 압축 파일 안의 `install.sh` 또는 `install.ps1` 이 `webreplica` 명령을 PATH에 설치해 줍니다.

## curl로 바로 설치

macOS / Linux에서는 아래 한 줄로 최신 릴리스를 자동 다운로드하고 설치할 수 있습니다.

```bash
curl -fsSL https://raw.githubusercontent.com/bssm-oss/web-replica/main/scripts/install-webreplica.sh | sh
```

설치 후에는 어느 폴더에서든 이렇게 실행하면 됩니다.

```bash
webreplica https://원하는사이트.com
```

주의: 이 도구는 원본 사이트를 픽셀 단위로 **완전히 똑같이 복제하는 도구가 아닙니다.** 기본값은 저작권/브랜드/추적 스크립트/장문 콘텐츠를 그대로 복사하지 않는 안전한 inspired reimplementation 입니다. 내 소유 사이트의 일부 에셋만 명시적으로 허용하려면 고급 옵션 `--allow-owned-assets` 를 사용하세요.

설치 위치를 직접 지정하고 싶으면:

```bash
curl -fsSL https://raw.githubusercontent.com/bssm-oss/web-replica/main/scripts/install-webreplica.sh | WEBREPLICA_INSTALL_DIR="$HOME/.local/bin" sh
```

## 요구 사항

- Go 1.26+
- Node.js
- npm
- Git
- Chrome 또는 Chromium
- 공식 Codex CLI

## 제일 쉬운 사용법

### 1. GitHub에서 바로 다운로드

Go를 몰라도 GitHub Releases에서 실행 파일을 바로 받을 수 있습니다.

1. [Releases 페이지](https://github.com/bssm-oss/web-replica/releases/latest)에 들어갑니다.
2. 내 운영체제에 맞는 파일을 받습니다.
   - Apple Silicon Mac: `webreplica_<version>_darwin_arm64.tar.gz`
   - Intel Mac: `webreplica_<version>_darwin_amd64.tar.gz`
   - Linux x64: `webreplica_<version>_linux_amd64.tar.gz`
   - Linux ARM64: `webreplica_<version>_linux_arm64.tar.gz`
   - Windows x64: `webreplica_<version>_windows_amd64.zip`
3. 압축을 풀고 `webreplica` 실행 파일을 터미널에서 실행합니다.

macOS / Linux 예시:

```bash
tar -xzf webreplica_<version>_darwin_arm64.tar.gz
cd webreplica_<version>_darwin_arm64
./install.sh
webreplica https://example.com
```

Windows 예시:

```powershell
Expand-Archive webreplica_<version>_windows_amd64.zip
cd webreplica_<version>_windows_amd64\webreplica_<version>_windows_amd64
.\install.ps1
webreplica https://example.com
```

`install.sh` 는 기본적으로 `/usr/local/bin/webreplica` 에 설치합니다. 권한이 필요하면 `sudo` 비밀번호를 물어볼 수 있습니다.
Windows의 `install.ps1` 은 `%USERPROFILE%\bin` 에 설치하고 사용자 PATH에 추가합니다. 새 터미널을 열면 `webreplica` 명령을 바로 쓸 수 있습니다.

### 2. Go로 설치

처음 설치는 한 번만 하면 됩니다.

Go가 이미 설치되어 있다면 이 한 줄이 제일 간단합니다.

```bash
go install github.com/bssm-oss/web-replica/cmd/webreplica@latest
```

소스 코드까지 받아서 설치하려면:

```bash
git clone https://github.com/bssm-oss/web-replica.git
cd web-replica
go mod tidy
go install ./cmd/webreplica
```

Codex CLI가 없다면 먼저 설치하세요.

```bash
npm i -g @openai/codex
codex
```

위 `codex` 명령으로 공식 로그인 절차를 진행하면 됩니다. 이 프로젝트는 `~/.codex/auth.json` 을 읽지 않으며, 토큰을 직접 처리하지 않습니다.

이제 아래처럼 링크만 넣으면 됩니다.

```bash
webreplica https://example.com
```

그러면 자동으로:

1. URL을 안전하게 검사합니다.
2. HTML과 화면 구조를 분석합니다.
3. desktop/tablet/mobile 스크린샷을 찍습니다.
4. `design-spec.json` 과 `brief.md` 를 만듭니다.
5. Codex CLI로 새 Vite + React + TypeScript + Tailwind 프로젝트를 생성합니다.
6. `npm install` 과 `npm run build` 를 실행합니다.
7. 결과를 `./generated-site` 폴더에 만듭니다.

실행 후 확인은 이렇게 합니다.

```bash
cd generated-site
npm run preview
```

또는 설치하지 않고 바로 실행하려면:

```bash
go run ./cmd/webreplica https://example.com
```

## 설치 / 고급 사용

기존 Siteforge 명령도 그대로 쓸 수 있습니다.

```bash
go install ./cmd/siteforge
```

## 빠른 시작

```bash
go run ./cmd/webreplica https://example.com
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

### `webreplica <url>`

가장 쉬운 기본 명령입니다. `siteforge build <url> --out ./generated-site --stack vite-react-tailwind` 와 같은 전체 파이프라인을 한 줄로 실행합니다.

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
- non-interactive `siteforge generate/build` 는 로컬 Codex MCP/플러그인 설정에 끌려가지 않도록 격리 옵션으로 실행합니다.

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
- Codex 자체 최종 요약 메시지가 실제 `npm install` / `npm run build` 결과와 다를 수 있으므로, Siteforge는 항상 후속 validator 실행 결과를 기준으로 성공/실패를 판정합니다.

## 향후 계획

- `next-tailwind` / `static-html-css` 실제 구현
- asset 다운로드 정책 세분화
- 시각적 diff 기반 검증 고도화
- 다중 페이지 분석
- 로컬 허용 대상(`--allow-local`) 명시 옵션 추가
