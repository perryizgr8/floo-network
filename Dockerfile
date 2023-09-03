FROM golang

# Set destination for COPY
WORKDIR /app

# Copy the source code. Note the slash at the end, as explained in
# https://docs.docker.com/engine/reference/builder/#copy
COPY *.go ./

# Download Go modules
COPY go.mod .
COPY go.sum .
RUN go mod tidy

# Build
RUN go build -o /server.go

# This is for documentation purposes only.
# To actually open the port, runtime parameters
# must be supplied to the docker command.
EXPOSE 1323

# Run
CMD [ "/floo-network" ]