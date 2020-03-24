##############################################################
## Stage 1 - Go Build
##############################################################

FROM golang:alpine AS builder

# install required
RUN apk update && apk add --no-cache bash git make

# Download source
RUN git clone https://github.com/cloud-barista/cb-tumblebug /go/src/github.com/cloud-barista/cb-tumblebug

# Dependencies download & Build
RUN cd /go/src/github.com/cloud-barista/cb-tumblebug && go mod download && \
    cd /go/src/github.com/cloud-barista/cb-tumblebug/src && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags '-w -extldflags "-static"' -o cb-tumblebug -v

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

COPY --from=builder /go/src/github.com/cloud-barista/cb-tumblebug/src/cb-tumblebug .
COPY --from=builder /go/src/github.com/cloud-barista/cb-tumblebug/conf ./conf
COPY ./run.sh .

# Sets the environment variables
RUN sed -i "s|$GOPATH/src/github.com/cloud-barista/cb-tumblebug|/app|g" ./conf/setup.env
RUN sed -i "s|localhost:1024|cb-spider:1024|g" ./conf/setup.env

CMD [ "./run.sh" ]

EXPOSE 1323

# docker build -t cb-tumblebug-image -f Dockerfile .
# docker run -it -p 1323:1323 cb-tumblebug-image