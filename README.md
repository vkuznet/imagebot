### imagebot
The imagebot service is designed to allow update of k8s services via CD/CI pipeline.
The CD/CI pipeline may be based on GitHub Action where one of the action will
call this service to request update of the image on k8s cluster.

To do that we need few pieces:
- add authorization token secret into your GitHub repository
- setup service on your favorie k8s infrastructure
- configure imagebot with allowed set of namespaces and services

#### imagebot configuration
The configuration of the service should include
```
# number of namespaces/services/repositories should be the same
# namespaces lists allowed namespaces
# services lists allowed services
# repository lists allowed repositories
{
    "port": 8111,
    "base": "",
    "token": "12345",
    "namespaces": ["ns1", "ns2"],
    "services": ["srv1", "srv2"],
    "repositories": ["repo1/srv1", "repo2/srv2"],
    "verbose": 1
}
```
For the full set of allowed parameters please see `server.go` code.

### repository settings
Please create appropriate GitHub secrets in your github repository:
- `IMAGEBOT_URL` represents image bot url
- `SERVICE_NAMESPACE` refers to application namespace
Please make sure that your namespace is part of imagebot configuration

On GitHub a package should add the following workflow steps:
```
# some workflow organization
name: Create issue on commit
on:
- push
jobs:
  create_commit:
    runs-on: ubuntu-latest
    steps:
    # step to add <------------ ADD THIS STEP
    - name: Post new image using REST API
      run: |
        curl --request POST \
		--url ${{ secrets.IMAGEBOT_URL }}
        --header 'authorization: Bearer ${{ secrets.GITHUB_TOKEN }}' \
        --header 'content-type: application/json' \
        --data '{
          "commit": "${{ github.sha }}",
		  "namespace": "${{ github.SERVICE_NAMESPACE }}",
		  "repository": "${{ github.repository }}",
		  "workflow": "${{ github.workflow }}"
          }'
```

### Testing
To test the imagebot workflow please set it up and run it, e.g.
```
# test the code
make test

# here is an example of config.json
cat config.json
{
    "port": 8111,
    "base": "",
    "token": "12345",
    "namespaces": ["foo", "bla"],
    "servics": ["srv1", "srv2"],
    "repositories": ["repo1/srv1", "repo2/srv2"],
    "verbose": 1
}

# run image bot
./imagebot -config config.json
```
Then, pleace a call to imageboth running on localhost
```
curl -v -X POST -H "Authorization: Bearer 12345" \
    -H "content-type: application/json" \
    -d '{"commit":"some-hash", "namespace": "ns", \
         "repository": "repo1/srv", "workflow": "test", "tag":"tag", "service":"srv"}' \
         http://localhost:8111
```
