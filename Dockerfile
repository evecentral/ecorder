FROM golang:latest 
RUN mkdir /app 
WORKDIR /app
RUN go get -v -u github.com/Masterminds/glide
ADD . /app/ 
RUN glide update && glide install
RUN go build -o ecorder ./ecorder/main.go
CMD ["/app/ecorder"]
