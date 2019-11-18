##############################################################
## Stage 1 - Go Build
##############################################################

FROM golang:alpine AS builder

# install required
RUN apk update && apk add --no-cache bash git gcc libc-dev

# set go environemnts
RUN go env -w GO111MODULE="off"

# downloads
RUN git clone https://github.com/cloud-barista/cb-spider /go/src/github.com/cloud-barista/cb-spider

# Dependencies download
RUN go get -u -v github.com/cloud-barista/cb-log
RUN go get -u -v github.com/cloud-barista/cb-store
RUN go get -u -v github.com/aws/aws-sdk-go/aws
RUN go get -u -v github.com/davecgh/go-spew/spew
RUN go get -u -v github.com/bramvdbogaerde/go-scp
RUN go get -u -v github.com/labstack/echo
RUN go get -u -v github.com/dgrijalva/jwt-go

# change setup.env path
RUN sed -i "s|~/go/src/github.com/cloud-barista/cb-spider|/go/src/github.com/cloud-barista/cb-spider|g" /go/src/github.com/cloud-barista/cb-spider/setup.env

# make dirver shared library
ENV CBSPIDER_ROOT=/go/src/github.com/cloud-barista/cb-spider
RUN cd /go/src/github.com/cloud-barista/cb-spider/cloud-control-manager/cloud-driver/drivers/aws && ./build_driver_lib.sh

# cb-spider build
RUN cd /go/src/github.com/cloud-barista/cb-spider/api-runtime/rest-runtime && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags '-w -extldflags "-static"' -o cb-spider -v

##############################################################
## Stage 2 - Applicaiton Setup
##############################################################

FROM ubuntu:latest

# use bash
RUN rm /bin/sh && ln -s /bin/bash /bin/sh

# set working directory and files
WORKDIR /app

COPY --from=builder /go/src/github.com/cloud-barista/cb-spider/api-runtime/rest-runtime/cb-spider .
COPY --from=builder /go/src/github.com/cloud-barista/cb-spider/cloud-driver-libs ./cloud-driver-libs
COPY --from=builder /go/src/github.com/cloud-barista/cb-spider/conf ./conf
COPY --from=builder /go/src/github.com/cloud-barista/cb-spider/setup.env .
COPY ./run.sh .

RUN sed -i "s|/go/src/github.com/cloud-barista/cb-spider|/app|g" ./setup.env

CMD [ "./run.sh" ]
#CMD [ "/bin/bash" ]

EXPOSE 1024