FROM golang:alpine

WORKDIR /app

COPY . ./

RUN apk update && apk add bash

RUN chmod a+x ./run.sh

CMD [ "./run.sh" ]