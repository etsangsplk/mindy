FROM eris/base:latest
MAINTAINER Eris Industries <support@erisindustries.com>

## Install mindy
RUN DEBIAN_FRONTEND=noninteractive apt-get update && apt-get dist-upgrade -y && apt-get install -y libgmp3-dev && apt-get clean all
ENV repo $GOPATH/src/github.com/eris-ltd/mindy
RUN mkdir -p $repo
COPY . $repo/
WORKDIR $repo
RUN go install

COPY start.sh /
RUN chmod 755 /start.sh

CMD ["/start.sh"]

