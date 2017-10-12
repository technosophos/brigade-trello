# Trello Gateway for Brigade

Trello API docs:


## Installation

Brigade is a prerequisite.


```
$ helm install -n brigade-trello charts/brigade-trello
```

You probably want to enable either Ingress or LoadBalancer

```
$ helm install -n brigade-trello charts/brigade-trello --set service.type=LoadBalancer
```

## Building

```
$ glide install
$ make build
```

or

```
$ make docker-build
```
