FROM apiaryio/dredd
EXPOSE 8080

RUN adduser go -h /go -s /bin/sh -D
RUN git clone https://gitlab.com/courtoisninja/cig-exchange-libs.git /go/src/cig-exchange-libs
RUN chown -fR go:go /go
RUN apk update && apk add go musl-dev git bash
USER go
ENV GOPATH /go
ENV GOBIN /go/bin
WORKDIR /go
COPY . /go/src/cig-exchange-homepage-backend
COPY .env /go

CMD sh -c 'go get .../ && go install cig-exchange-homepage-backend'
