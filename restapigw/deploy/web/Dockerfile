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
ADD reflex .
ADD entrypoint.sh .
#ADD ./conf/hmac.yaml ./conf

# fetch dependencies
#RUN go mod tidy
#RUN go mod download
#RUN go get -d -v ./...

# build the binary
RUN CGO_ENABLED=0 go build -o ./hmac_server -v ./...
RUN chmod +x ./hmac_server
RUN chmod +x ./reflex
RUN chmod +x ./entrypoint.sh

#RUN rm -rf ./Dockerfile ./go.mod ./makefile

#####################################################
##### STEP 2. Build a small image (only executable)
#####################################################

# FROM scratch

# # copy static executable
# COPY --from=builder /app/hmac_server .
# COPY --from=builder /app/reflex .



# run the binary
#ENTRYPOINT [ "/app/reflex -g '/app/conf/hmac.yaml' -s -- /app/hmac_server" ]
#CMD [ "./reflex -s -- ./hmac_server" ]
#ENTRYPOINT [ "./hmac_server" ]
ENTRYPOINT [ "./entrypoint.sh" ]

EXPOSE 8010