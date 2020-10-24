FROM golang:alpine

WORKDIR /app
COPY . /app

RUN go build -o app

CMD ["/app/app"]