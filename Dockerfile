FROM golang:1.14.2-alpine3.11 as builder

RUN mkdir /build
ADD . /build/

WORKDIR /build

RUN CGO_ENABLED=0 GOOS=linux go build -o silent-assassin  ./cmd/silent-assassin/*.go

FROM golang:1.14.2-alpine3.11

COPY --from=builder /build/silent-assassin /layers/golang/app/

WORKDIR /layers/golang/app/

CMD ["sh", "-c", " ./silent-assassin start server"]
