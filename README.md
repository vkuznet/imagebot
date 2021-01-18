### imagebot

[![GitHubActions Status](https://github.com/vkuznet/imagebot/workflows/Build/badge.svg)](https://github.com/vkuznet/imagebot/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/vkuznet/imagebot)](https://goreportcard.com/report/github.com/vkuznet/imagebot)

The imagebot service is designed to update k8s services via CD/CI pipeline.
The CD/CI pipeline may be based on GitHub Action where one of the action will
call this service to request update of the image on k8s infrastructure.

To do that we need few pieces:
- add secrets into your GitHub repository
- setup service on your k8s infrastructure
- configure imagebot with allowed set of namespaces, repositories and services

#### Required secrets
In order to enabled CD/CI pipeline for your github repository and allow
imagebot to update images on k8s infrastructure
please create appropriate GitHub secrets in your github repository:
- `IMAGEBOT_URL` represents URL of imagebot service
- `SERVICE_NAMESPACE` refers to application namespace
- `DOCKER_IMAGE` referts to docker hub image name

Please note: your namespace should be part of imagebot configuration

#### imagebot configuration
The configuration of the service should include
```
# secret represents imagebot secret used for token creation/encryption
# tokenInterval setups validity interval in seconds for generated token
# namespaces lists allowed namespaces
# services lists allowed services
# repository lists allowed repositories
# please note: namespaces/services/repositories list represent triplets
{
    "port": 8111,
    "base": "",
    "namespaces": ["ns1", "ns2"],
    "services": ["srv1", "srv2"],
    "images": ["repo1/srv1", "repo2/srv2"],
    "secret": "bla-bla-bla",
    "tokenInverval": 600,
    "verbose": 1
}
```
For the full set of allowed parameters please see `server.go` code.

### Testing procedure
To test the service please run it as following:
```
imagebot -config config.json
```
and place two requests:
```
# first request to obtain the token
curl -v -X POST -H "content-type: application/json" \
    -d '{"commit":"dad51084ea82ab2f6f573b6daa464ed0d7c23a1d", "namespace": "http", \
         "repository": "vkuznet/httpgo", "tag":"00.00.01", "image": "httpgo:123", \
         "service":"httpgo"}' http://localhost:8111/token


# second request to process workflow
curl -v -X POST -H "Authorization: Bearer $token" -H "content-type: application/json" \
    -d '{"commit":"dad51084ea82ab2f6f573b6daa464ed0d7c23a1d", "namespace": "http", \
         "repository": "vkuznet/httpgo", "tag":"00.00.01", "image": "httpgo:123", \
         "service":"httpgo"}' http://localhost:8111
```

### repository settings
GitHub package should have the following workflow steps:
```
# some workflow organization
name: Create issue on commit
on:
- push
jobs:
  create_commit:
    runs-on: ubuntu-latest
    steps:
    #
    ####### ADD THIS STEP to your GitHubAction workflow
    #
    - name: Post new image using REST API
      run: |
        curl --request POST \
        --url ${{ secrets.IMAGEBOT_URL }}/token \
        --header 'content-type: application/json' \
        --data '{
          "commit": "${{ github.sha }}",
          "namespace": "${{ github.SERVICE_NAMESPACE }}",
          "repository": "${{ github.repository }}",
          "tag": <repo>/<package>:${{steps.get-ref.outputs.tag}},
          "image": "${{ github.DOCKER_IMAGE }}",
          "service": "<NAME_OF_YOUR_SERVICE>"
          }'
        curl --request POST \
        --url ${{ secrets.IMAGEBOT_URL }} \
        --header 'authorization: Bearer ${{ secrets.GITHUB_TOKEN }}' \
        --header 'content-type: application/json' \
        --data '{
          "commit": "${{ github.sha }}",
          "namespace": "${{ github.SERVICE_NAMESPACE }}",
          "repository": "${{ github.repository }}",
          "tag": <repo>/<package>:${{steps.get-ref.outputs.tag}},
          "image": "${{ github.DOCKER_IMAGE }}",
          "service": "<NAME_OF_YOUR_SERVICE>"
          }'
```
