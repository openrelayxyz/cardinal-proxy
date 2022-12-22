# Cardinal Proxy

Cardinal Proxy is intended for use in conjunction with
[Cardinal EVM](https://github.com/openrelayxyz/cardinal-evm) and [Flume](https://github.com/openrelayxyz/cardinal-flume).

Cardinal EVM and Flume combined provide the majority of standard Web3 RPC
methods, plus some extended features. Web3 applications, however, generally
expect all of these methods to be available from the same HTTP endpoint.
Cardinal Proxy can be used to route requests to the appropriate backend
services.

## Usage

Cardinal Proxy can be run with:

```
./cardinal-proxy config.yaml
```

The provided config.yaml should work, modifying the various backend URLs to
point to appropriate service endpoints.


## Alpha Warning

Cardinal Proxy is under early development. It is not in production use at
Rivet, and should be used with caution.

## Roadmap

Now that Cardinal Proxy has basic functionality to route to Cardinal EVM and
Flume, the next roadmap item is a plugin framework. This will allow us (and
you!) to build customizable functionality. Some examples of plugins we hope to
build within the plugin framework:

* ETH Web3 subscriptions (eg. newHeads, pendingTransactions, logs, etc.)
* A caching framework
* Network-specific capabilities for networks with additional RPC methods
