## Node

The node controls the server it is installed on.

The node has the ability to install and configure external software (e.g. *Docker* and *iptables*) [].

Clients communicate with the node through the *sync service*.

Nodes communicate with eachother through the *gossip service*.

Nodes have the following immutable properties:
   - Public key
   - Address
   - Gossip service port
   - Sync service port

A node is uniquely identified by its public key, and no other nodes can use the same key-pair. The node resource identifier is formed by hashing the public key bytes with Blake2b-128, and encoding the hash using Bech32 with a prefix (`node`).

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

### Test mode

The node has a test mode for unit testing many of its features locally. While testing, only a local directory is used.

| Path                                                            | Description               |
| --------------------------------------------------------------- | ------------------------- |
| `$TEST_DIR/<node-id>/assets/<asset-content-hash>`               | Storage per node          |
| `$TEST_DIR/<node-id>/functions/<function-id>/[0-9]+`            | Function workspaces       |
| `$TEST_DIR/<node-id>/functions/<function-id>/[0-9]+/handler.js` | Function handlers         |
| `$TEST_DIR/<node-id>/key`                                       | Node Ed25519 private key  |
| `$TEST_DIR/<node-id>/ledger`                                    | Test project ledger       |
| `$TEST_DIR/<node-id>/logs/<resource-id>/<yyyy/mm/dd-hh:mm:ss>`  | Logs created by resources |

### Common paths

   - `ASSETS_DIR`
   - `FUNCTIONS_WORKSPACE`
   - `LEDGER_PATH`
   - `LOGS_DIR`
   - `PRIVATE_KEY_PATH`