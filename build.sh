#!/usr/bin/env bash

sync(){
    rsync -avzh --delete --exclude=".git,.github" ./* root@45.76.208.186:/tmp/horus-operator/
}

build(){
    cd /tmp/horus-operator
    docker build -t hybfkuf/horus-operator:latest .
    docker push hybfkuf/horus-operator:latest
}


sync
ssh root@45.76.208.186 "
    $(typeset -f build)
    build
"
