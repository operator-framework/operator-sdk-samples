ARG BASEIMAGE
FROM ${BASEIMAGE}
USER 0

# Ensure fresh metadata rather than cached metadata in the base by running
# yum clean all && rm -rf /var/yum/cache/* first
RUN yum clean all && rm -rf /var/cache/yum/* \
 && yum -y update \ 
 && yum install -y python36-devel gcc libffi-devel

# ubi7
RUN if $(cat /etc/redhat-release | grep --quiet 'release 7'); then (yum install -y python36-pip || true); fi
# ubi8
RUN if $(cat /etc/redhat-release | grep --quiet 'release 8'); then (yum install -y python3-pip || true); fi

RUN pip3 install --user molecule==2.22

ARG NAMESPACEDMAN
ADD $NAMESPACEDMAN /namespaced.yaml
ADD build/test-framework/ansible-test.sh /ansible-test.sh
RUN chmod +x /ansible-test.sh
USER 1001
ADD . /opt/ansible/project
