##############################################################
## Stage 1 - Go Build
##############################################################

FROM golang:alpine AS builder

RUN apk update && apk add --no-cache bash

WORKDIR /app

ADD ./cmd ./cmd
ADD ./conf ./conf
ADD ./pkg ./pkg
ADD ./go.mod .
ADD ./main.go .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags '-w -extldflags "-static"' -tags cb-restapigw -o cb-restapigw -v

#############################################################
## Stage 2 - Applicaiton Setup
##############################################################

FROM ubuntu:latest

# use bash
RUN rm /bin/sh && ln -s /bin/bash /bin/sh

WORKDIR /app

COPY --from=builder /app/cb-restapigw .
COPY --from=builder /app/conf ./conf

ENTRYPOINT [ "/app/cb-restapigw", "-c", "/app/conf/cb-restapigw.yaml", "-d", "run" ]

EXPOSE 8000
