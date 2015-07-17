# Start from a Debian image with the latest version of Go installed
# and a workspace (GOPATH) configured at /go.
FROM golang

# Copy the local package files to the container's workspace.
ADD . /go/src/github.com/bradberger/mqtt-client-logger

# Build the outyet command inside the container.
# (You may fetch or manage dependencies here,
# either manually or with a tool like "godep".)
RUN go get git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git
RUN go get gopkg.in/inconshreveable/log15.v2
RUN go get gopkg.in/gorp.v1
RUN go get github.com/go-sql-driver/mysql
RUN go get gopkg.in/gorp.v1
RUN go install github.com/bradberger/mqtt-client-logger

# Run the outyet command by default when the container starts.
ENTRYPOINT /go/bin/mqtt-client-logger

# Document that the service listens on port 8080.
EXPOSE 3000
