FROM golang:latest 
RUN mkdir -p /go/src/github.com/evecentral/ecorder
WORKDIR /go/src/github.com/evecentral/ecorder
RUN go get -v -u github.com/Masterminds/glide
ADD . /go/src/github.com/evecentral/ecorder
RUN glide update && glide install
RUN go build -o ecorder-bin ./ecorder/main.go
CMD ["/go/src/github.com/evecentral/ecorder/ecorder-bin"]
