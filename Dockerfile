FROM ubuntu:latest
LABEL authors="selholper"

ENTRYPOINT ["top", "-b"]