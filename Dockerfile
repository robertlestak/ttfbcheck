FROM golang:1.16-alpine as builder

WORKDIR /src

COPY . .

RUN go build -o ttfbcheck *.go

FROM golang:1.16-alpine as app

COPY --from=builder /src/ttfbcheck /usr/local/bin/

ENTRYPOINT [ "ttfbcheck" ]