##############################################################
## Stage 1 - Go Build
##############################################################

FROM golang:alpine AS builder

<<<<<<< HEAD:restapigw/deploy-test/Dockerfile
RUN apk update && apk add --no-cache bash
=======
RUN apk update && apk add --no-cache bash git && apk --no-cache add ca-certificates
>>>>>>> develop:restapigw/deploy/cb-restapigw/Dockerfile

WORKDIR /app

<<<<<<< HEAD:restapigw/deploy-test/Dockerfile
ADD ../cmd/ ./cmd/
ADD ../conf/ ./conf/
ADD ../pkg/ ./pkg/
ADD ../go.mod .
ADD ../main.go .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags '-w -extldflags "-static"' -tags cb-restapigw -o cb-restapigw -v
=======
# Copy sources
WORKDIR /app
COPY ./src . 

# build source
RUN cd /app && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags '-w -extldflags "-static"' -tags cb-restapigw -o cb-restapigw -v
>>>>>>> develop:restapigw/deploy/cb-restapigw/Dockerfile

#############################################################
## Stage 2 - Applicaiton Setup
##############################################################

FROM ubuntu:latest

# use bash
RUN rm /bin/sh && ln -s /bin/bash /bin/sh

WORKDIR /app

<<<<<<< HEAD:restapigw/deploy-test/Dockerfile
COPY --from=builder /app/cb-restapigw /app

ENTRYPOINT [ "/app/cb-restapigw", "-c", "/app/conf/cb-restapigw.yaml", "-d", "run" ]

EXPOSE 8000
=======
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/cb-restapigw .
COPY --from=builder /app/conf ./conf

ENTRYPOINT [ "/app/cb-restapigw", "-c", "/app/conf/etri.yaml", "-d", "run" ]

EXPOSE 8000

# docker build -t cb-restapigw-image -f Dockerfile .
# docker run -p 8000:8000 -it cb-restapigw-image
>>>>>>> develop:restapigw/deploy/cb-restapigw/Dockerfile
