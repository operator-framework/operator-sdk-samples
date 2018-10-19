ARG BASEIMAGE
FROM ${BASEIMAGE}
ADD build/_output/bin/memcached-operator-test /usr/local/bin/memcached-operator-test
ARG NAMESPACEDMAN
ADD $NAMESPACEDMAN /namespaced.yaml
ADD build/test-framework/go-test.sh /go-test.sh
