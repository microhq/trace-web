FROM alpine:3.2
ADD templates /templates
ADD trace-web /trace-web
WORKDIR /
ENTRYPOINT [ "/trace-web" ]
