FROM alpine:3.6

RUN adduser -D dex-operator
USER dex-operator

ADD tmp/_output/bin/dex-operator /usr/local/bin/dex-operator
