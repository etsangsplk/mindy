# Pull base image.
FROM eris/base

ENV DOCKERFILES ./DOCKER/linked

# grab deps (gmp)
RUN apt-get update && \
  apt-get install -y --no-install-recommends \
    libgmp3-dev dnsutils && \
  rm -rf /var/lib/apt/lists/*

# install client tools and eris-keys
ENV REPO github.com/eris-ltd/mint-client 
RUN git clone https://$REPO $GOPATH/src/$REPO
WORKDIR $GOPATH/src/$REPO
RUN git checkout staging
RUN go install ./...

RUN go get github.com/eris-ltd/eris-keys

# copy in the priv_validator
ADD ./test/plumbing/priv_validator.json /priv_validator.json
RUN chown $USER:$USER /priv_validator.json

ADD ./test/plumbing/test.sh /test.sh
RUN chown $USER:$USER /test.sh

# set user
USER $USER
WORKDIR /home/eris

CMD ["/test.sh"]
