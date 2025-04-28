FROM golang:1.22.0-alpine3.18 AS builder

RUN apk add upx

# Set the working directory outside $GOPATH to enable the support for modules.
WORKDIR /src

# Copy the local package files to the container's workspace.
ADD . /src

# Build the command inside the container.
RUN GOOS=linux GOARCH=mipsle GOMIPS=softfloat go build -ldflags "-w -s" -o thornol_app
# compress the binary
RUN upx --best --lzma thornol_app