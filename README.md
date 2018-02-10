# Trello Gateway for Brigade

EXPERIMENTAL: This is a gateway for exposing Trello-compatible webhooks to
Brigade. It is not ready for production use (namely, the payload validation
is not done yet).

To use this gateway, you will need:

- A Kubernetes cluster
  - running Brigade
  - with Helm
- A Trello account, along with the developer token and apikey
- At least one Trello board

Trello API docs: https://developers.trello.com/

## Installation

Clone this repo and install the Helm chart:


```
$ helm install -n brigade-trello charts/brigade-trello
```

You probably want to enable either Ingress or LoadBalancer

```
$ helm install -n brigade-trello charts/brigade-trello --set service.type=LoadBalancer
```

You may also want to enabled RBAC with `--set rbac.enabled=true`.

Once the gateway is installed and exposed on a public IP address, you will need
to register a Webhook gateway with Trello. This is a manual process
described in the [Trello docs](https://developers.trello.com/page/webhooks).

The webhook URL should take the form:

```
http://<GATEWAY_IP_OR_NAME>/trello/<PROJECT_ID>
```
(substitute `https` if your ingress or service is configured for TLS)

Note that the `PROJECT_ID` is of the form `brigade-XXXXXXXXXXXXXXXXXXXX`, not the
human-readable name.

We highly recommend registering a webhook for _a single board_ until you get
the hang of using this gateway.

There is a highly experimental "generic webhook" implementation in this gateway
as well. You can use this to run arbitrary webhooks, but it is insecure by
design (since it does not seek to validate the identy of the sender or the
payload). It is at `http://<HOST>/webhook/<PROJECT_ID>`.

> Note: While the above URLs respond to GET and HEAD requests, it will always
> respond with a no-op 200-level request. Only POST can be used to actually
> trigger an event.

## Writing a Brigade File

This gateway emits two Brigade events:

- `trello`: The event triggered any time a Trello event fires.
- `webhook`: (EXPERIMENTAL) Triggered if you use the experimental `/generic/<PROJECT>`
  endpoint. This hook is considered _untrusted_ and should be treated with
  utmost caution.

A `brigade.js` that makes use of the `trello` hook can be found in the repository.
In a nutshell, though, they look like this:

```javascript
events.on("trello", (e, p) => {
  // Parse the JSON payload from Trello.
  var hook = JSON.parse(e.payload)

  // Now you can go about your business.
})
```

The Trello JSON format is described [in their docs](https://developers.trello.com/page/webhooks)

## Building from source

```
$ glide install
$ make build
```

or

```
$ glide install
$ make docker-build
```
