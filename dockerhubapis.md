### DockerHub APIs
```
# get list of images
curl -k "https://hub.docker.com/v2/repositories/cmssw/?page_size=100" | jq

# get registry token
export TOKEN=$(curl -kfsSL "https://auth.docker.io/token?service=$AUTH_SERVICE&scope=$AUTH_SCOPE" | jq --raw-output '.token')

# get list of tags for given image
curl -kfsSL -H "Authorization: Bearer $TOKEN" "https://registry.hub.docker.com/v2/cmssw/auth-proxy-server/tags/list" | jq
```
