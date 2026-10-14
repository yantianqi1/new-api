FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /sft-collector .

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata
COPY --from=build /sft-collector /usr/local/bin/sft-collector
WORKDIR /data/sft-collector
ENTRYPOINT ["/usr/local/bin/sft-collector"]
