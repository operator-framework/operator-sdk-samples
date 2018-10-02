FROM alpine:3.6

RUN adduser -D memcached-operator
USER memcached-operator

ADD build/_output/bin/memcached-operator /usr/local/bin/memcached-operator
