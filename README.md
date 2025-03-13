# Open Web Services

## Introduction

Open Web Services (OWS) is an **Open Source alternative for** some of **AWS**'s managed services:

| Service              | AWS equivalent   | Implemented |
| -------------------- | ---------------- | ----------- |
| Permissions          | IAM              | [x] (MVP)   |
| Serverless compute   | Lambda Functions | [x] (MVP)   |
| REST APIs            | API Gateway      | [x] (MVP)   |
| Database tables      | DynamoDB         | [ ]         |
| Storage              | S3               | [ ]         |
| Events               | EventBridge      | [ ]         |
| DNS                  | Route53          | [ ]         |
| CDN                  | CloudFront       | [ ]         |
| Workflows            | Step Functions   | [ ]         |
| Private repositories | CodeCommit       | [ ]         |
| CI/CD                | CodePipeline     | [ ]         |
| ...                  |                  |             |

To achieve AWS-like resilience, OWS uses a **P2P network of private servers**.
Each participating server runs the OWS *node*, which hosts the managed services and syncs server state with peers. **All peers are equal** and never assume master roles.

The OWS *client* queries node state and submits infrastructure configuration changes, called *change sets*.
Change sets are signed by the client, and are only accepted by the nodes if the signer has the necessary permissions.
The OWS client can derive change sets from JSON files [], allowing OWS nodes to be configured using [Infrastructure as Code (IaC)](https://en.wikipedia.org/wiki/Infrastructure_as_code) [].

### Table of contents

**User documentation**:

  - [Download](https://christianschmitz.github.io/ows/index.html)

**Specification**:

   1. [P2P Network](./doc/specification/01-P2P_network.md)
   2. [Ledger](./doc/specification/02-Ledger.md)
   3. [Node](./doc/specification/03-Node.md)
   4. [Client](./doc/specification/04-Client.md)

### Motivation

AWS is one of the best cloud hosting providers:
   - High up-times
   - Innovative solutions for event-driven micro-service architectures

However, AWS has the following notable disadvantages:
   - High-end EC2 servers are more expensive than competitors
   - Using AWS-specific services makes migrating to another cloud provider very difficult

The latter disadvantage has recently been amplified by global political uncertainty, especially for companies using AWS datacenters abroad.

OWS uses only standard cloud technologies, so it can be used with any cloud provider (or any combination of cloud providers). Migrating an OWS project to another cloud provider is seemless.

### Contributing

Replicating all AWS managed services is a huge amount of work, which is only feasible if multiple companies, who want to migrate away from AWS, collaborate.

OWS is Open Source, with a BSD-3-Clause license.

The node and client are written in Golang.

The client hosts a React [Single Page Application (SPA)](https://en.wikipedia.org/wiki/Single-page_application), inspired by the AWS console, and written in Typescript [].

OWS includes a Typescript library for IaC, replacing the AWS CDK [].
