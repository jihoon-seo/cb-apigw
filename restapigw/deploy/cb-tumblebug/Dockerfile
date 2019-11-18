##############################################################
## Stage 1 - Go Build
##############################################################

FROM golang:alpine AS builder

# install required
RUN apk update && apk add --no-cache bash git make

# Dependencies download
RUN go get -u -v github.com/cloud-barista/cb-log
RUN go get -u -v github.com/cloud-barista/cb-store

# MCISM download
RUN go get -u -v github.com/cloud-barista/cb-tumblebug/...

# MCISM build
RUN cd /go/src/github.com/cloud-barista/cb-tumblebug/src && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags '-w -extldflags "-static"' -o mcism -v

##############################################################
## Stage 2 - Applicaiton Setup
##############################################################

FROM ubuntu:latest

# use bash
RUN rm /bin/sh && ln -s /bin/bash /bin/sh

# apt-get source change for speed
RUN sed -i "s|archive.ubuntu.com|mirror.kakao.com|g" /etc/apt/sources.list

# install requires
RUN apt-get update && apt-get install -y apt-utils etcd-server

# set working directory and files
WORKDIR /app

COPY --from=builder /go/src/github.com/cloud-barista/cb-tumblebug/src/mcism .
COPY --from=builder /go/src/github.com/cloud-barista/cb-tumblebug/conf ./conf
COPY ./run.sh .

RUN sed -i "s|~/go/src/github.com/cloud-barista/cb-tumblebug|/app|g" ./conf/setup.env
RUN sed -i "s|localhost:1024|cb-spider:1024|g" ./conf/setup.env

CMD [ "./run.sh" ]
#CMD [ "/bin/bash" ]

EXPOSE 1323