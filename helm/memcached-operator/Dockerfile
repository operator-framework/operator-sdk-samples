# Build the manager binary
FROM quay.io/operator-framework/helm-operator:v1.0.0-alpha.2

ENV HOME=/opt/helm
COPY watches.yaml ${HOME}/watches.yaml
COPY helm-charts  ${HOME}/helm-charts
WORKDIR ${HOME}
