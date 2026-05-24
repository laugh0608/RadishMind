FROM golang:1.22-bookworm AS build

WORKDIR /src/services/platform

COPY services/platform/go.mod ./
RUN go mod download

COPY services/platform/ ./

ARG BUILD_VERSION=local
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags "-X main.buildVersion=${BUILD_VERSION}" \
    -o /out/radishmind-platform \
    ./cmd/radishmind-platform

FROM python:3.12-slim AS runtime

ENV PYTHONUNBUFFERED=1 \
    RADISHMIND_PLATFORM_LISTEN_ADDR=0.0.0.0:7000 \
    RADISHMIND_PLATFORM_PROVIDER=mock \
    RADISHMIND_PLATFORM_MODEL=radishmind-local-docker \
    RADISHMIND_PLATFORM_PYTHON_BIN=python3 \
    RADISHMIND_PLATFORM_BRIDGE_SCRIPT=scripts/run-platform-bridge.py

WORKDIR /workspace

RUN pip install --no-cache-dir "jsonschema>=4.0"

COPY --from=build /out/radishmind-platform /usr/local/bin/radishmind-platform
COPY services/__init__.py services/__init__.py
COPY services/gateway/ services/gateway/
COPY services/runtime/ services/runtime/
COPY scripts/run-platform-bridge.py scripts/run-platform-bridge.py
COPY contracts/ contracts/
COPY prompts/ prompts/

EXPOSE 7000

CMD ["radishmind-platform"]
