## Build
FROM golang:1.19-alpine AS builder
RUN go env -w GO111MODULE=auto \
  && go env -w GOPROXY=https://goproxy.cn,direct  \
  && sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && cat /etc/apk/repositories \
  && apk add --no-cache bash git openssh
WORKDIR /build

COPY ./ .
RUN go build -o /Blog2gin


## Deploy
FROM alpine:latest
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && cat /etc/apk/repositories
WORKDIR /app
RUN apk add --no-cache tzdata curl
COPY ./ .
COPY --from=builder /Blog2gin /app/Blog2gin

RUN chmod +x /app/Blog2gin
ENTRYPOINT ["/app/Blog2gin"]