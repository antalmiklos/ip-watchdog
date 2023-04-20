FROM golang:1.20

WORKDIR /app

COPY . .
RUN go mod download && go mod verify

RUN go build -v -o /usr/local/bin/app ./...

CMD ["app"]