FROM golang:latest 
RUN mkdir /app 
ADD . /app/ 
WORKDIR /app 
RUN go build -o ecorder ./ecorder/main.go
CMD ["/app/ecorder"]
