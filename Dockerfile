FROM alpine:3.8
MAINTAINER Sten Røkke <sten.ivar.rokke@nav.no>
WORKDIR /app

COPY uptimed .

CMD /app/uptimed --logtostderr=true
