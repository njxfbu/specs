---
title: "Retrieval Provider (Miner)"
---

# Provider Dependencies

The Retrieval Provider Depends On The Following Dependencies

- **Host**: A libp2p host (set setup the libp2p protocols)
- **Filecoin Node**: A node implementation to query the chain for pieces and to setup and manage payment channels
- **StorageMining Subsystem**: For unsealing sectors
- **BlockStore**: Same as one used by data transfer module
- **Data Transfer**: V1 only -- Module used for transferring payload. Reads from the blockstore.

# API

{{< readfile file="retrieval_provider.id" code="true" lang="go" >}}
