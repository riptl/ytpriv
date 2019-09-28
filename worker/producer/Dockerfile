# Build stage
FROM golang:alpine as builder
# Build stage dependencies
RUN apk --no-cache add git
RUN apk --no-cache add build-base


# Build librdkafka
RUN apk --no-cache add bash pkgconf
WORKDIR /deps/librdkafka
RUN git clone https://github.com/edenhill/librdkafka /deps/librdkafka
RUN ./configure --prefix=/usr \
    --enable-static \
    --disable-zstd \
    --disable-ssl \
    --disable-sasl \
    --disable-lz4-ext \
    --disable-lz4
RUN make
RUN make install

WORKDIR /app/mango

# Get root dependencies
ADD go.* ./
RUN go mod download

# Get producer dependencies
ADD worker/producer/go.* ./worker/producer/
RUN cd worker/producer && go mod download

# Build app
ADD . /app/mango
RUN cd worker/producer \
 && go build \
    -installsuffix cgo \
    -ldflags "-s -w -linkmode external -extldflags -static" \
    -o /go/bin/producer \
    -tags static_all \
    .

# Light stage
FROM scratch
COPY --from=builder /go/bin/producer /bin/producer
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
CMD ["/bin/producer"]
EXPOSE 8000
ENV KAFKA_CONFIG=/config \
    KAFKA_TOPIC=comments \
    KAFKA_CLOSE_TIMEOUT=3000 \
    HTTP_LISTEN=:8000
