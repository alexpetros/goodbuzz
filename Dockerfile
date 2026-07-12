# ---------- Builder ----------
FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

RUN go install github.com/a-h/templ/cmd/templ@v0.2.747

COPY . .

RUN /go/bin/templ generate

RUN CGO_ENABLED=0 GOOS=linux go build -o goodbuzz

# ---------- Runtime ----------
FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache sqlite ca-certificates

COPY --from=builder /app/goodbuzz .
COPY --from=builder /app/router ./router
COPY --from=builder /app/db ./db

COPY entrypoint.sh .
RUN chmod +x entrypoint.sh

ENV GOODBUZZ_PORT=8080

EXPOSE 8080

ENTRYPOINT ["./entrypoint.sh"]