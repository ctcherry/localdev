FROM debian:stretch

RUN apt-get update -y -qq \
&& apt-get install -y -qq \
dnsmasq \
&& rm -rf /var/lib/apt/lists/*

COPY dnsmasq.conf /etc/dnsmasq.conf

CMD ["/usr/sbin/dnsmasq", "-d", "--log-facility=-", "--log-queries=extra"]
