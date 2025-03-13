## Ledger

The project configuration is called the *ledger*.

The ledger is a blockchain-like linked-list. The first entry is the initial project configuration, and the subsequent entries are the change sets.

Each ledger entry consists of *actions* and *signatures*.

The ledger uses CBOR encoding for its binary representation.

### Initial configuration

The initial configuration consists primarily of the `SetSyncPort`, `SetGossipPort` and `AddNode` actions.

The signers of the initial configuration are given unrevokeable root permissions, and can submit any future change set.

The initial configuration can optionally specify the following:
   - Minimal root user quorum needed for future change sets []
   - Public blockchain smart contract address containing valid root user public keys []

The project's identifier is the Bech32 encoded hash of the initial configuration. For example: `project1lum5n8xamyqappegr2xgkgaeeqnty6xj`.

### Change set

Each change set must refer to the hash of the previous change set (or hash of the initial configuration), and is signed by any number of OWS clients.

Change sets contain a list of *actions*. Each action requires specific permissions. A change set is rejected if none of its signatures has a public key with the required permissions.

Usually change sets are only signed by a single client, but multiple signatures can be included to accomodate [group signature](https://en.wikipedia.org/wiki/Group_signature) permissions [].

Further change set validation consists of ensuring:
   - Uniqueness of resource names and port numbers
   - Attributes of added resources are within bounds
   - Modified resource attributes are within bounds
   - New resource references exist
   - Modified resource references exist
   - Referenced resources aren't removed

### Actions

OWS defines the following ledger actions:

   - AddFunction
   - AddGateway
   - AddGatewayEndpoint
   - AddNode
   - AddUser
   - RemoveFunction
   - RemoveGateway
   - RemoveGatewayEndpoint []
   - RemoveNode []
   - RemoveUser []
   - ...

### Resource identifiers

Every newly created resource is given a deterministic identifier. The identifier is calculated as follows:
   1. Take the hash of the previous change set (empty bytes in case the resource is defined in the initial configuration)
   2. Take the little endian encoding of the action index
   3. Hash the concatenation of the bytes from step 1 and 2 using [Blake2b](https://en.wikipedia.org/wiki/BLAKE_(hash_function))-128
   4. Generate a string with human-readable prefix using [Bech32](https://en.bitcoin.it/wiki/Bech32). A different prefix is used for each resource type

An example of a resource identifier is `gateway1syxs9p6s3497f7we5lzjp3a9csyx4402`.

Unlike AWS's ARNs, OWS resource identifiers don't contain custom names. This way resource names can be changed without impacting the resource identifier. For convenience, the client maps custom resource names to resource identifiers when querying the project state and when submitting change sets [].

OWS resource identifiers are globally unique, just like AWS's ARN's

### Permissions

Similar to AWS's IAM, OWS uses *policies* containing *policy statements* for fine-grained permission management.

A policy statement consists of:
   - List of resource identifiers
   - List of actions (format: `<category>:<action-name>`)
   - Effect ("Allow" or "Deny")

The wildcard symbol (`*`) can be used to match all actions, all actions of a specific category, and/or all resource identifiers.

Actions that create resources, don't operate on existing resources. Such actions are instead considered to operate on a generic global resource, identified by `*`. This means that actions that create resources must also use a wildcard in the resource identifiers list for a positive match.

A policy consists of multiple policy statements. To determine the change set action permissions, use the following steps:
   1. The allowed actions are collected in a set
   2. The denied actions are removed from this set
   3. If the requested action isn't mentioned in the resulting set, the action is rejected
   
The order of policy statements in the policy doesn't matter.