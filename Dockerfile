#FROM alpine:3.7
FROM golang:1.12.7
ADD build build
ADD build/idhub /idhub
RUN chmod u+x /idhub

CMD ["/idhub", "-logtostderr=true", "-stderrthreshold=INFO"]

