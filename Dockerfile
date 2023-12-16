FROM golang:alpine AS builder

WORKDIR /build

ADD src .
ADD www .

COPY . .

RUN go build -o wggo main.go

FROM alpine

WORKDIR /app

COPY --from=builder /build/wggo /app/wggo
#COPY --from=builder /build/config.yml /app/config.yml
COPY --from=builder /build/www /app/www

CMD ["/app/wggo"]