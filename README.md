### 📘 Explore Service (Backend Candidate Exercise - Summer 2024)
#### 🧠 Overview

This project implements a subset of Muzz’s production Explore Service, which manages user interactions (likes/passes) in the matching system.
It’s built with Go, gRPC, and PostgreSQL, and follows clean, modular architecture.

#### 🏗️ Architecture

```
explore-service/
├── cmd/
│   └── main.go                # gRPC server entrypoint
├── internal/
│   ├── server/                # gRPC service implementation
│   │   └── explore.go
│   ├── repository/            # Database queries and persistence
│   │   └── decision_repo.go
│   └── models/                # Domain models (if needed)
├── proto/
│   └── explore.proto          # Protocol buffer definition
├── Dockerfile
├── docker-compose.yml
└── README.md
```

#### ⚙️ Technology Stack
|Component |Purpose |
|---|---|
|Go (1.23) | Backend language|
|gRPC + Protobuf | High-performance RPC communication|
|PostgreSQL | Persistent data store|
|Docker / Docker Compose | Containerization and local orchestration|

🚀 Features

- Implements 4 RPC endpoints:

  - PutDecision — Record or update a user’s decision (like/pass)
  - ListLikedYou — List all users who liked a given user
  - ListNewLikedYou — List users who liked you but you haven’t liked back
  - CountLikedYou — Count how many users liked a given user

- Existing decisions can be overwritten

- Scales efficiently for users with hundreds of thousands of decisions

- Fully containerized for easy testing and deployment

🧩 Database Schema

```sql
CREATE TABLE IF NOT EXISTS decisions (
actor_user_id TEXT NOT NULL,
recipient_user_id TEXT NOT NULL,
liked_recipient BOOLEAN NOT NULL,
created_at TIMESTAMPTZ DEFAULT NOW(),
updated_at TIMESTAMPTZ DEFAULT NOW(),
PRIMARY KEY (actor_user_id, recipient_user_id)
);

CREATE INDEX IF NOT EXISTS idx_recipient_user_id ON decisions (recipient_user_id);
CREATE INDEX IF NOT EXISTS idx_actor_user_id ON decisions (actor_user_id);
```

#### 🐳 Run with Docker Compose

```bash
docker-compose up --build
```

This will start:

- A PostgreSQL database (port `5432`)

- The Explore gRPC service (port `50051`)

⚡ Run Locally (without Docker)

Start PostgreSQL:

```bash
docker run -d --name explore-db -p 5432:5432 \
 -e POSTGRES_USER=postgres \
 -e POSTGRES_PASSWORD=postgres \
 -e POSTGRES_DB=explore \
 postgres:16
```

Run the Go service:

```bash
go run cmd/main.go
```

Server runs at:

```
localhost:50051
```

#### 🧪 Example gRPC Calls

1️⃣ PutDecision

```
grpcurl -plaintext \
 -d '{"actor_user_id":"alice","recipient_user_id":"bob","liked_recipient":true}' \
 localhost:50051 explore.ExploreService/PutDecision
```

2️⃣ ListLikedYou

```
grpcurl -plaintext \
 -d '{"recipient_user_id":"bob"}' \
 localhost:50051 explore.ExploreService/ListLikedYou
```

3️⃣ ListNewLikedYou

```
grpcurl -plaintext \
 -d '{"recipient_user_id":"bob"}' \
 localhost:50051 explore.ExploreService/ListNewLikedYou
```

4️⃣ CountLikedYou

```
grpcurl -plaintext \
 -d '{"recipient_user_id":"bob"}' \
 localhost:50051 explore.ExploreService/CountLikedYou
```

#### 🧱 Scaling Considerations

- Primary key (actor_user_id, recipient_user_id) prevents duplicates and simplifies overwrites.

- Indexes on recipient_user_id and actor_user_id improve query speed for high-volume users.

- All queries are optimized with indexed lookups and minimal joins.

- Stateless gRPC service — easy to scale horizontally with load balancers.

- PostgreSQL connection pooling can be managed by pgbouncer or a similar proxy.

#### 🧰 Testing

Run all tests:

```bash
go test ./... -v
```

Example unit tests cover:

- Mutual likes detection

- Correct overwrite behavior

- Query correctness for ListNewLikedYou and CountLikedYou

#### 🧾 Assumptions

- Users are represented only by their IDs (no profiles).

- Mutual like means both users have liked_recipient = true towards each other.

- Pagination is omitted for simplicity (can be added using OFFSET/LIMIT or tokens).

- No authentication — internal microservice-level access only.

#### 👨‍💻 Author

Adem Şahin
Backend Engineer — Go, gRPC, Kubernetes, Distributed Systems
LinkedIn | GitHub
# grpc-example
