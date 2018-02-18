SHELL := /bin/bash

all: fedora ubuntu;

fedora:
	docker build -t fmind/fedora images/fedora

ubuntu:
	docker build -t fmind/ubuntu images/ubuntu
