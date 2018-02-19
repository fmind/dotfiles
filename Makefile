SHELL := /bin/bash

all: fedora;

fedora:
	docker build -t fmind/fedora images/fedora
