FROM golang:latest 
RUN mkdir /app 
ADD . /app/ 
WORKDIR /app
RUN go get -v -u github.com/Masterminds/glide
RUN glide update && glide install
RUN go build -o ecorder ./ecorder/main.go
CMD ["/app/ecorder"]
