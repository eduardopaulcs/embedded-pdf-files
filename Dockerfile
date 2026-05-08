# Base
FROM golang:1.26.2-alpine3.23 AS base

WORKDIR /app
ENV PATH="/root/go/bin:${PATH}"

COPY go.mod go.sum* ./
RUN go mod download

# Dev
FROM base AS dev

RUN go install github.com/air-verse/air@v1.65.1

COPY . .

# Prod
FROM base AS builder

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/app .

FROM alpine:3.23 AS prod

COPY --from=builder /bin/app /bin/app
CMD ["/bin/app"]
