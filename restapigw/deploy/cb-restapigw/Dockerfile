##############################################################
## Stage 1 - Go Build
##############################################################

FROM golang:alpine AS builder

RUN apk update && apk add --no-cache bash git 

# set go environemnts
RUN go env -w GO111MODULE=""

# download source
#RUN git clone https://github.com/cloud-barista/cb-apigw.git /app
RUN git clone https://github.com/ccambomorris/cb-apigw.git /app

# build source
RUN cd /app/restapigw && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags '-w -extldflags "-static"' -tags cb-restapigw -o cb-restapigw -v

#############################################################
## Stage 2 - Applicaiton Setup
##############################################################

FROM ubuntu:latest

# use bash
RUN rm /bin/sh && ln -s /bin/bash /bin/sh

WORKDIR /app

COPY --from=builder /app/restapigw/cb-restapigw /app

ENTRYPOINT [ "/app/cb-restapigw", "-c", "/app/conf/cb-restapigw.yaml", "-d", "run" ]
#CMD [ "/bin/bash" ]

EXPOSE 8000