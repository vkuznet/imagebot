#!/bin/sh
token=`curl -v -s -X POST -H "content-type: application/json" -d '{"commit":"COMMIT", "namespace": "NAMESPACE", "repository": "REPOSITORY", "image": "IMAGE", "tag":"TAG", "service":"SERVICE"}' HOST/token`
echo "TOKEN: $token"
curl -v -s -X POST -H "Authorization: Bearer $token" -H "content-type: application/json" -d '{"commit":"COMMIT", "namespace": "NAMESPACE", "repository": "REPOSITORY", "image": "IMAGE",  "tag":"TAG", "service":"SERVICE"}' HOST/
