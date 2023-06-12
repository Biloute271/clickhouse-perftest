FROM golang:latest as builder
LABEL maintainer="Stanislas Dmitriev <stanislas.dmitriev@gmail.com>"
WORKDIR /app
COPY . .
COPY config.yaml .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .
FROM alpine:latest  
RUN apk --no-cache add ca-certificates
WORKDIR /bin/
COPY --from=builder /app/main .
COPY --from=builder /app/config.yaml .
EXPOSE 8080
CMD ./main 