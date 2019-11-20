#####################################################
##### STEP 1. Build executable binary
#####################################################

FROM golang:alpine AS builder 

# install git for fetching the dependencies
RUN apk update && apk add --no-cache git

# make folder and copy sources
RUN mkdir -p /app
RUN mkdir -p /app/conf

WORKDIR /app
ADD go.mod .
ADD main.go .
#ADD entrypoint.sh .

# build the binary
RUN CGO_ENABLED=0 go build -o ./hmac_server -v ./...
RUN chmod +x ./hmac_server
#RUN chmod +x ./entrypoint.sh

#ENTRYPOINT [ "./entrypoint.sh" ]
ENTRYPOINT [ "./hmac_server" ]

EXPOSE 8010