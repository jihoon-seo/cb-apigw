##############################################################
## Stage 1 - Go Build
##############################################################

FROM golang:alpine AS builder

# Install requires
RUN apk update && apk add --no-cache bash git gcc libc-dev

# Download source
RUN git clone https://github.com/cloud-barista/cb-spider /go/src/github.com/cloud-barista/cb-spider

# Dependencies download
RUN cd /go/src/github.com/cloud-barista/cb-spider && go mod download

# Make driver shared library
ENV CBSPIDER_ROOT=/go/src/github.com/cloud-barista/cb-spider
RUN cd /go/src/github.com/cloud-barista/cb-spider/cloud-control-manager/cloud-driver/drivers/aws && ./build_driver_lib.sh

# Build cb-spider
RUN cd /go/src/github.com/cloud-barista/cb-spider/api-runtime/rest-runtime && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags '-w -extldflags "-static"' -o cb-spider -v

##############################################################
## Stage 2 - Applicaiton Setup
##############################################################    

FROM ubuntu:latest

# Set working directory and files
WORKDIR /app

COPY --from=builder /go/src/github.com/cloud-barista/cb-spider/api-runtime/rest-runtime/cb-spider .
COPY --from=builder /go/src/github.com/cloud-barista/cb-spider/cloud-driver-libs ./cloud-driver-libs
COPY --from=builder /go/src/github.com/cloud-barista/cb-spider/conf ./conf 
COPY --from=builder /go/src/github.com/cloud-barista/cb-spider/setup.env .

# Change setup.env
RUN sed -i "s|$GOPATH/src/github.com/cloud-barista/cb-spider|/app|g" ./setup.env

COPY ./run.sh .

CMD [ "./run.sh" ]

EXPOSE 1024

# docker build -t cb-spider-image -f Dockerfile .
# docker run -p 1024:1024 -it cb-spider-image