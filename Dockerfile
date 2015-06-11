FROM eris/base
MAINTAINER Eris Industries <support@erisindustries.com>

## Install djbdns
## complications on debian due to licensing (?)
RUN DEBIAN_FRONTEND=noninteractive apt-get update && apt-get dist-upgrade -y && apt-get install -y make daemontools daemontools-run ucspi-tcp procps multitail && apt-get clean all
ADD http://http.us.debian.org/debian/pool/main/d/djbdns/dbndns_1.05-8_amd64.deb /tmp/dbndns_1.05-8_amd64.deb
RUN dpkg -i /tmp/dbndns_1.05-8_amd64.deb

## Configure tinydns users
RUN useradd -s /bin/false tinydns && \
    useradd -s /bin/false dnslog

## Config dir and service
RUN tinydns-conf tinydns dnslog /etc/tinydns 0.0.0.0
RUN cd /etc/service && ln -sf /etc/tinydns

## Install mindy
RUN apt-get install -y libgmp3-dev
ENV repo $GOPATH/src/github.com/eris-ltd/mindy
RUN mkdir -p $repo
COPY . $repo/
WORKDIR $repo
RUN go install

COPY start.sh /
RUN chmod 755 /start.sh

EXPOSE 53/udp 
EXPOSE 53

CMD ["/start.sh"]

