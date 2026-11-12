# syntax=docker/dockerfile:1

ARG GO_VERSION="1.27"
ARG XX_VERSION="1.10.0"
ARG ALPINE_VERSION="3.24"
ARG GOLANGCI_LINT_VERSION="v2.15.0"
ARG GOLANGCI_FROM_SOURCE="true"
ARG GOPLS_VERSION="v0.24.0"

FROM --platform=$BUILDPLATFORM tonistiigi/xx:${XX_VERSION} AS xx

FROM --platform=$BUILDPLATFORM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS base
ENV GOFLAGS="-buildvcs=false"
RUN apk add --no-cache gcc linux-headers musl-dev
COPY --from=xx --link / /
WORKDIR /src

FROM base AS golangci-build
ARG GOLANGCI_LINT_VERSION
ADD "https://github.com/golangci/golangci-lint.git#${GOLANGCI_LINT_VERSION}" .
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/ go mod download
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/ mkdir -p out && go build -o /out/golangci-lint ./cmd/golangci-lint

FROM --platform=$BUILDPLATFORM golangci/golangci-lint:${GOLANGCI_LINT_VERSION}-alpine AS golangci-lint
FROM scratch AS golangci-binary-false
COPY --from=golangci-lint /usr/bin/golangci-lint golangci-lint
FROM scratch AS golangci-binary-true
COPY --from=golangci-build /out/golangci-lint golangci-lint
FROM golangci-binary-${GOLANGCI_FROM_SOURCE} AS golangci-binary

FROM base AS gopls-build
ARG GOPLS_VERSION
ADD "https://github.com/golang/tools.git#gopls/${GOPLS_VERSION}" /tools
WORKDIR /tools/gopls
COPY <<'EOF' internal/cmd/ftpgrab-analyzers/main.go
package main

import (
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/gopls/internal/analysis/embeddirective"
	"golang.org/x/tools/gopls/internal/analysis/fillreturns"
	"golang.org/x/tools/gopls/internal/analysis/infertypeargs"
	"golang.org/x/tools/gopls/internal/analysis/maprange"
	"golang.org/x/tools/gopls/internal/analysis/nonewvars"
	"golang.org/x/tools/gopls/internal/analysis/noresultvalues"
	"golang.org/x/tools/gopls/internal/analysis/simplifycompositelit"
	"golang.org/x/tools/gopls/internal/analysis/simplifyslice"
	"golang.org/x/tools/gopls/internal/analysis/unusedparams"
	"golang.org/x/tools/gopls/internal/analysis/yield"
)

func main() {
	multichecker.Main(
		embeddirective.Analyzer,
		fillreturns.Analyzer,
		infertypeargs.Analyzer,
		maprange.Analyzer,
		nonewvars.Analyzer,
		noresultvalues.Analyzer,
		simplifycompositelit.Analyzer,
		simplifyslice.Analyzer,
		unusedparams.Analyzer,
		yield.Analyzer,
	)
}
EOF
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/ \
  go build -o /out/gopls-analyzers ./internal/cmd/ftpgrab-analyzers

FROM base AS gopls-analyze
ARG TARGETPLATFORM
RUN --mount=type=bind,target=. \
    --mount=type=cache,target=/root/.cache,id=gopls-cache-$TARGETPLATFORM \
    --mount=from=gopls-build,source=/out/gopls-analyzers,target=/usr/bin/gopls-analyzers \
  xx-go --wrap && \
  go vet -mod=vendor -vettool=/usr/bin/gopls-analyzers ./...

FROM base AS lint
ARG TARGETPLATFORM
RUN --mount=type=bind,target=. \
    --mount=type=cache,target=/root/.cache,id=lint-cache-$TARGETPLATFORM \
    --mount=from=golangci-binary,source=/golangci-lint,target=/usr/bin/golangci-lint \
  xx-go --wrap && \
  golangci-lint run ./...
