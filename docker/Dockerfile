FROM golang:1.12.5 AS builder
#FROM golang:latest

MAINTAINER Guilherme Martins


RUN apt-get update && apt-get install -y net-tools git sudo vim strace wget curl bzip2 procps cron logrotate \
g++ libprotobuf-c-dev libprotobuf-c1 libprotobuf-dev libgeoip-dev flex bison nginx links

#MD5 (dep-linux-amd64) = 84c5a8828203699d6a515b827471f75b
#SHA1 c35010206dfa3735effdd888a4c3bf9d68775cf4  opt_influxdb_core.tgz
#SHA1 f760ee4fcccbcb9a79455f913e93f286083faa8c  libprotobuf7_2.4.1-3_amd64.deb 

ADD https://github.com/golang/dep/releases/download/v0.5.0/dep-linux-amd64 /usr/bin/dep
#COPY ./dep-linux-amd64 /usr/bin/dep
RUN chmod +x /usr/bin/dep
COPY libprotobuf7_2.4.1-3_amd64.deb /home/root/
COPY opt_influxdb_core.tgz /home/root/
RUN dpkg -i /home/root/libprotobuf7_2.4.1-3_amd64.deb 
RUN useradd -ms /bin/bash encore
USER encore
WORKDIR /home/encore
RUN mkdir /home/encore/go/
ENV GOPATH=/home/encore/go/

RUN go get github.com/ggmartins/encore
RUN cd $GOPATH/src/github.com/ggmartins/encore/ && dep ensure -vendor-only && \
go build -o $GOPATH/bin/encore -ldflags "-X main.gitRevisionId=$(git rev-parse HEAD)" *.go && \
go build -o $GOPATH/bin/encore-parser $GOPATH/src/github.com/ggmartins/encore/encore-parser/parser.go
#RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix nocgo -o /app .
RUN ln -s /usr/share/GeoIP/GeoIP.dat /home/encore/GeoIP.dat
USER root
RUN tar xvzf /home/root/opt_influxdb_core.tgz -C /
RUN ln -s $GOPATH/src/github.com/ggmartins/encore/contrib/debian-init-script /etc/init.d/encore
RUN ln -s $GOPATH/src/github.com/ggmartins/encore/contrib/cron.hourly /etc/cron.hourly/encore
RUN ln -s $GOPATH/src/github.com/ggmartins/encore/contrib/logrotate-config /etc/logrotate.d/encore
RUN ln -s /opt/influxdb/current/scripts/init.sh /etc/init.d/influxdb
RUN mkdir /var/log/encore && chown encore:encore /var/log/encore/
#RUN mkdir -p /srv/encore/config && chown encore:encore -R /srv/encore
RUN update-rc.d encore defaults
RUN update-rc.d influxdb defaults
USER encore
#VOLUME ['/opt/influxdb/shared', '/srv/']
EXPOSE 80 443 8890-8899
