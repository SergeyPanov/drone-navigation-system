FROM golang:1.18
ENV DNS_CONFIG=/etc/drone-navigation-system/dns.yaml
WORKDIR /drone-navigation-system
COPY . /drone-navigation-system
RUN mkdir -p /etc/drone-navigation-system/
RUN mv -n dns.yaml /etc/drone-navigation-system/dns.yaml
RUN make build
EXPOSE 8080
ENTRYPOINT ["make", "run"]
