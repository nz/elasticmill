#!/bin/bash
while true; do
  curl -s -o /dev/null -X POST http://localhost:5000/test/test/`openssl rand -base64 10 | tr -cd '[:alnum:]'` \
    --data-binary "{now:\"`date`\", rand:\"`openssl rand -base64 10`\"}" && echo -n . || true
done
