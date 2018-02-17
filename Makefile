SHELL := /bin/bash

all: fedora ubuntu debian;

fedora:
	docker build -t fmind/fedora images/fedora

ubuntu:
	docker build -t fmind/ubuntu images/ubuntu

debian:
	docker build -t fmind/debian images/debian
