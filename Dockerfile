FROM golang:1.17 as Builder
WORKDIR /ronak/src

# we copy the root
COPY ./ ./

# Compile the code
WORKDIR /ronak/src
RUN ls -la
RUN go build -a -ldflags '-s -w' -o /ronak/bin/dns-resolver ./


FROM ubuntu:20.04
WORKDIR /ronak/bin
RUN apt-get update
RUN apt install -y ca-certificates && update-ca-certificates

WORKDIR /ronak/bin
COPY --from=Builder /ronak/bin/dns-resolver ./

# Entry point
ENTRYPOINT ["/ronak/bin/dns-resolver"]
