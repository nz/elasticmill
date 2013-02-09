#!/bin/bash

url=$1
if [ "$url" = "" ]; then
  url="http://localhost:9200/"
fi

# Send over a bunch of random documents
while true; do
  curl -s -o /dev/null -X POST "$url/test/test/`openssl rand -base64 10 | tr -cd '[:alnum:]'`" \
    --data-binary "{now:\"`date`\", rand:\"`openssl rand -base64 10`\"}" && echo -n . || true
done
