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
{
    "port": 8111,
    "base": "",
    "token": "12345",
    "namespaces": ["abc", "zys"],
    "servics": ["srv1", "srv2"],
    "verbose": 1
}
```
For the full set of allowed parameters please see `server.go` code.
