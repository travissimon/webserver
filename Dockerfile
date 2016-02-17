FROM scratch
MAINTAINER travis.simon@nicta.com.au

ADD webserver webserver

ENV PORT 8080
EXPOSE 8080
ENTRYPOINT ["/webserver"]