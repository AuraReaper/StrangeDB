## v2.0.0 (2025-12-23)

### Feat

- integrate hinted handoff and antientroppy to the cluster node
- **storage**: tombstone mangement
- **antientropy**: implement merkele tree
- **coordinator**: implement hinted handoff
- **coordinator**: implement read repair for get

## v1.0.0 (2025-12-21)

### Feat

- **coordinator**: implement read/write coordinator
- **gossip**: implemnet core gossip protocol
- **grpc**: implement grpc client and server
- **ring**: implement consistent hash ring
- **grpc**: protobuf defination and setup

### Refactor

- integrate ring, gossiper, coordinator with node
- solve endoing issues and typo mistakes

## v0.2.0 (2025-12-21)

### Feat

- **node**: node lifecycle managemnt
- **http**: create http api server
- **storage**: storage layer implementation
- **storage**: badgerdb implemntation
- **hlc**: implement hlc
- **logger**: setup zerlog logger
- **config**: configurtion load and setup

### Fix

- **Makefile**: update makefile to support latest go version

### Refactor

- update gitignore
- **storage**: correct return type of Delete function
