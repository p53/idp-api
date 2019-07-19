FROM golang:1.12-alpine3.9 as builder

WORKDIR /go/src/github.com/p53/idp-api
ENV GOPATH=/go

RUN set -ex \
    && apk add --no-cache --virtual .build-deps \
            git \
            gcc \
            g++ \
            make \
            libc-dev \
            musl-dev \
            linux-headers \
            curl
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
COPY . /go/src/github.com/p53/idp-api

RUN dep ensure
RUN go build

FROM alpine:3.9

COPY --from=builder /go/src/github.com/p53/idp-api/idp-api /bin/idp-api
COPY --from=builder /go/src/github.com/p53/idp-api/swagger.yml /

USER nobody

ENTRYPOINT ["/bin/idp-api"]
