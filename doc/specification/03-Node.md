## Node

The node controls the server it is installed on.

The node should have the ability to install and configure external software (e.g. *Docker* and *iptables*).

Clients communicate with the node through the OWS *API service*.

Nodes communicate with each other through the OWS *gossip service*.

Nodes have the following immutable properties:
   - Public key
   - Address
   - API service port
   - Gossip service port

A node is uniquely identified by its public key, and no other nodes can use the same key-pair. The node resource identifier is formed by hashing the public key bytes with Blake2b-128, and encoding the hash using Bech32 with the `node` prefix.

### Custom HTTPS

The API service and gossip service potentially handle sensitive data that must be hidden from middle-men.
It is reasonable to use HTTPS as the communication protocol.

Instead of using certificates signed by a CA, certificates can be derived directly from the Ed25519 keys assigned to each client and node. The HTTPS certificate validation mechanism must then be modified to extract the public keys from client/server certificates, and match those public keys against a whitelist derived from the ledger.

### Files

The node persists its data using the following file structure:

| Path                                                     | Description                         |
| -------------------------------------------------------- | ----------------------------------- |
| `/etc/init.d/ows`                                        | OWS daemon controller               |
| `/etc/ows/key`                                           | Node Ed25519 private key            |
| `/usr/bin/ows`                                           | Node binary                         |
| `/var/lib/ows/assets/<asset-content-hash>`               | General storage location            |
| `/var/lib/ows/functions/<function-id>/[0-9]+`            | Function workspaces                 |
| `/var/lib/ows/functions/<function-id>/[0-9]+/handler.js` | Function handlers                   |
| `/var/lib/ows/ledger`                                    | Project ledger                      |
| `/var/log/ows/<resource-id>/<yyyy/mm/dd-hh:mm:ss>`       | Logs created by resources           |

Unlike the client, the node doesn't support multiple projects. A node is intended to run for a single project only.

### Detached mode

The node is controlled by init.d and runs in the background.

### Test mode

The node has a test mode for unit testing its features locally. While testing, only a local directory is used.

| Path                                                            | Description               |
| --------------------------------------------------------------- | ------------------------- |
| `$TEST_DIR/<node-id>/assets/<asset-content-hash>`               | Storage per node          |
| `$TEST_DIR/<node-id>/functions/<function-id>/[0-9]+`            | Function workspaces       |
| `$TEST_DIR/<node-id>/functions/<function-id>/[0-9]+/handler.js` | Function handlers         |
| `$TEST_DIR/<node-id>/key`                                       | Node Ed25519 private key  |
| `$TEST_DIR/<node-id>/ledger`                                    | Test project ledger       |
| `$TEST_DIR/<node-id>/logs/<resource-id>/<yyyy/mm/dd-hh:mm:ss>`  | Logs created by resources |

### Asset existence signing

Some services depend on assets. For example, serverless functions depend on their handler assets. Though each node should be able to run each serverless function defined in a project, it doesn't need to persist the underlying handler asset.

Nodes are however not able to approve change sets that depend on assets they are unaware of. Therefore asset existence must be communicated between the nodes.