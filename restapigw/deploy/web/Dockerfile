#####################################################
##### STEP 1. Build executable binary
#####################################################

FROM golang:alpine AS builder 

# install git for fetching the dependencies
RUN apk update && apk add --no-cache git

WORKDIR /app

ADD go.mod .
ADD main.go .

# download dependencies
RUN go get github.com/labstack/echo/v4 && \
  go get github.com/spf13/viper && \
  # build the binary
  CGO_ENABLED=0 go build -o ./hmac_server -v ./... && \
  chmod +x ./hmac_server

##############################################################
## Stage 2 - Applicaiton Setup
##############################################################

FROM scratch

# set working directory and files
WORKDIR /app

COPY --from=builder /app/hmac_server .

ENTRYPOINT [ "/app/hmac_server" ]

EXPOSE 8010