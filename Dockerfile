FROM golang:1.21.5

WORKDIR /usr/src/app

COPY . .
RUN go build -v -o /usr/local/bin/app ./...

CMD ["app"]
