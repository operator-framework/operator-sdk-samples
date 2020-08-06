source ~/.virtualenvs/molecule/bin/activate

cd ../memcached-operator
export IMG='quay.io/asmacdo/example-operator:v0.0.1'
export OPERATOR_IMAGE=${IMG}
make docker-build docker-push
time molecule test -s kind
