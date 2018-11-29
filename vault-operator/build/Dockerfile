FROM alpine:3.6

ADD build/_output/bin/vault-operator /usr/local/bin/vault-operator

RUN adduser -D vault-operator
USER vault-operator
