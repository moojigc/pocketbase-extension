FROM golang:1.19

WORKDIR /app

COPY . /app/

RUN go build .
RUN mv pb.chimid.rocks /usr/local/bin/pocketbase
RUN chmod +x /usr/local/bin/pocketbase

# Notify Docker that the container wants to expose a port.
EXPOSE 8091

# Start Pocketbase
CMD [ "/usr/local/bin/pocketbase", "serve", "--http", "0.0.0.0:8090" ]