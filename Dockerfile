FROM ubuntu:16.04

RUN apt-get update -qq
RUN DEBIAN_FRONTEND=noninteractive apt-get -y install golang
RUN DEBIAN_FRONTEND=noninteractive apt-get -y install git