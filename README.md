# Open Web Services

## Introduction

Open Web Services (OWS) is an Open Source alternative for some popular managed services offered by major cloud providers:

| Service              | AWS equivalent   | Azure equivalent | GCS equivalent            | Status |
| -------------------- | ---------------- | ---------------- | ------------------------- | ------ |
| Permissions          | IAM              | RBAC             | IAM                       | MVP    |
| Serverless compute   | Lambda Functions | Azure Functions  | Cloud Functions           | MVP    |
| REST APIs            | API Gateway      | API Management   | API Gateway               | MVP    |
| Database tables      | DynamoDB         | Cosmos DB        | Firestore                 | Todo   |
| Storage              | S3               | Blob Storage     | Cloud Storage             | Todo   |
| Events               | EventBridge      | Event Grid       | Eventarc                  | Todo   |
| DNS                  | Route53          | Azure DNS        | Cloud DNS                 | Todo   |
| CDN                  | CloudFront       | Front Door       | Cloud CDN                 | Todo   |
| Workflows            | Step Functions   | Logic Apps       | Workflows                 | Todo   |
| Private repositories | CodeCommit       | Azure Repos      | Cloud Source Repositories | Todo   |
| CI/CD                | CodePipeline     | Azure Pipelines  | Cloud Build               | Todo   |
| ...                  |                  |                  |                           |        |

To achieve high up-times and resilience, OWS uses a **P2P network of private servers**.
Each participating server runs the OWS *node*, which hosts the managed services and syncs server state with peers. **All peers are equal** and never assume master roles.

The OWS *client* queries node state and submits infrastructure configuration changes, called *change sets*.
Change sets are signed by the client, and are only accepted by the nodes if the signer has the necessary permissions.

**User documentation**:

  - [Download](https://christianschmitz.github.io/ows/index.html)

**Architecture**:

   1. [P2P Network](./doc/specification/01-P2P_network.md)
   2. [Ledger](./doc/specification/02-Ledger.md)
   3. [Node](./doc/specification/03-Node.md)
   4. [Client](./doc/specification/04-Client.md)

### Motivation

The major cloud providers have the following advantages:
   - High up-times
   - Innovative solutions for event-driven micro-service architectures

But using them comes with the following disadvantages:
   - High-end servers are more expensive than smaller competitors
   - Vendor lock-in: using vendor-specific managed services makes migrating to another cloud provider very difficult

The latter disadvantage has recently been amplified by global political uncertainty, especially for companies using foreign datacenters.

OWS runs on standard technologies, so it can be used with any cloud provider (or any combination of cloud providers). Migrating an OWS project to another cloud provider is seemless.

Nice-to-haves:
   - Instantaneous and atomic deployments
   - Ability to test complete infrastructure deployments locally

**Why not Terraform?**

Terraform is a cloud-agnostic IaC language and deployment tool, it however does not offer cloud-agnostic abstractions of eg. serverless functions.
### Contributing

Emulating all popular managed services is a huge amount of work, which is only feasible if multiple entities collaborate.

OWS is Open Source, with a BSD-3-Clause license.

The node and client are written in Golang.

**Todo**:
   - Implement the OWS equivalent of each popular managed service
   - Ability to derive change sets from JSON files, allowing OWS nodes to be configured using [Infrastructure as Code (IaC)](https://en.wikipedia.org/wiki/Infrastructure_as_code)
   - A React [Single Page Application (SPA)](https://en.wikipedia.org/wiki/Single-page_application) hosted by the client, inspired by the AWS console, and written in Typescript
   - A Typescript library for IaC, inspired by the declarative AWS Typescript CDK
   - Advanced change set validation, based on git commit signatures or other signatures
   - The node automatically installs install itself and creates a `/etc/init.d/ows` script when running in detached mode
   - Client binaries compatible with Windows and MacOS
   - ...
