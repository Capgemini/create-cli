FROM golang:1.20-alpine AS base
WORKDIR /app
COPY . ./
RUN go mod download
RUN CGO_ENABLED=0 go build -o create-cli .

FROM alpine:3.18.2  
RUN apk add --no-cache tini
WORKDIR /root/
COPY --from=base /app/create-cli ./
ENTRYPOINT ["/sbin/tini", "--", "./create-cli"]
