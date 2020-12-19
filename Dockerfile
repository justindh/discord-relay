FROM golang:1.14

WORKDIR gosniff/src
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

VOLUME /var/gosniff/
ENTRYPOINT [ "GoSniff", "-p" ]