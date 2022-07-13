FROM golang:1.18
ARG DNS_CONFIG_ARGUMENT=./dns.yaml
ARG PORT=8080
ENV DNS_CONFIG=/etc/drone-navigation-system/dns.yaml
WORKDIR /drone-navigation-system
COPY ["*.go", "Makefile", "go.mod", "go.sum", "/drone-navigation-system/"]
COPY ${DNS_CONFIG_ARGUMENT} /etc/drone-navigation-system/dns.yaml
RUN make build
EXPOSE $PORT
ENTRYPOINT ["make", "run"]
