from golang:1.13-alpine as build

WORKDIR /go/src/app
COPY . .
RUN go build -v -o /app .

################################

from alpine:3.11
COPY --from=build /app /app

ENV HEADER_SECRET=CHANGE_ME
ENV OUTPUT=/generated
ENV PORTRAITS=/portraits
ENV INTERVAL=60
ENV CARPET=http://localhost:3141

COPY ./run.sh .

RUN chmod +x run.sh

CMD [ "sh", "-c", "./run.sh"]