#!/bin/sh
host=http://localhost:8111
echo "obtain token"
token=`curl -s -X POST -H "content-type: application/json" -d '{"commit":"dad51084ea82ab2f6f573b6daa464ed0d7c23a1d", "namespace": "http", "repository": "vkuznet/httpgo", "image": "cmssw/httpgo", "tag":"00.00.01", "service":"httpgo"}' $host/token`
echo "new token $token"
echo ""
echo "should fail with wrong token"
curl -s -X POST -H "Authorization: Bearer xxx$token" -H "content-type: application/json" -d '{"commit":"dad51084ea82ab2f6f573b6daa464ed0d7c23a1d", "namespace": "http", "repository": "vkuznet/httpgo", "image": "cmssw/httpgo", "tag":"00.00.01", "service":"httpgo"}' $host
echo ""
echo "should fail to process request"
curl -s -X POST -H "Authorization: Bearer $token" -H "content-type: application/json" -d '{"commit":"dad51084ea82ab2f6f573b6daa464ed0d7c23a1d", "namespace": "http", "repository": "vkuznet/httpgo", "image": "cmssw/httpgo", "tag":"00.00.01", "service":"httpgo"}' $host
