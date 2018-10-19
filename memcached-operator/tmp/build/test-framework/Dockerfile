ARG BASEIMAGE

FROM ${BASEIMAGE}

ADD tmp/_output/bin/memcached-operator-test /usr/local/bin/memcached-operator-test
ARG NAMESPACEDMAN
ADD $NAMESPACEDMAN /namespaced.yaml
ADD tmp/build/go-test.sh /go-test.sh
