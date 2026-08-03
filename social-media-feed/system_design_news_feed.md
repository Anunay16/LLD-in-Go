# System Design Guide: Facebook-like News Feed at Scale

Welcome! This guide explains how to design a high-scale social media feed system (like Facebook or Twitter/X) from scratch. 

We will break down every concept step-by-step, starting from basic principles and building up to how tech giants handle millions of users in real time. **No code or complex Low-Level Design (LLD) — just clear concepts, analogies, and architecture.**

---

## Table of Contents
1. [Core Analogy & System Overview](#1-core-analogy--system-overview)
2. [Requirements Breakdown & Scale Estimation](#2-requirements-breakdown--scale-estimation)
3. [Architecture: Monolith vs. Microservices](#3-architecture-monolith-vs-microservices)
4. [Service Boundaries & Responsibilities](#4-service-boundaries--responsibilities)
5. [API Design (High Level)](#5-api-design-high-level)
6. [Deep Dive 1: Feed Generation (Push vs. Pull vs. Hybrid)](#6-deep-dive-1-feed-generation-push-vs-pull-vs-hybrid)
7. [Deep Dive 2: Database Strategy (SQL vs. NoSQL)](#7-deep-dive-2-database-strategy-sql-vs-nosql)
8. [Deep Dive 3: Real-Time Likes & Views (High Concurrency)](#8-deep-dive-3-real-time-likes--views-high-concurrency)
9. [Deep Dive 4: Caching Strategy & Fast Feed Load](#9-deep-dive-4-caching-strategy--fast-feed-load)
10. [Solving the Celebrity / Viral Content Problem](#10-solving-the-celebrity--viral-content-problem)
11. [Summary Checklist for Interviews](#11-summary-checklist-for-interviews)

---

## 1. Core Analogy & System Overview

### The Newspaper Analogy
Imagine a traditional newspaper:
- **Publishing (Writing)**: A journalist writes an article. It gets printed once.
- **Reading**: Millions of people buy the paper and read the exact same front page.

Now imagine a **Personalized Newspaper** (Facebook Feed):
- Every single reader has a **custom front page** containing updates only from their friends and pages they follow.
- If you have 500 friends, your feed is created by blending content from all 500 friends in reverse chronological order (or ranked by relevance).
- When a friend posts a photo, your personal front page needs to update almost instantly.

### The Big Challenge
1. **High Read-to-Write Ratio**: Users read feeds far more often than they post (e.g., 100 reads for every 1 post).
2. **Massive Fan-out**: If Cristiano Ronaldo (with 100M followers) posts a video, that single post must reach 100M custom newspapers simultaneously.
3. **Instant Latency**: Users expect their feed to load in under 200 milliseconds.

---

## 2. Requirements Breakdown & Scale Estimation

### Functional Requirements (What the system does)
1. **Create Post**: Users can publish posts with text, images, or videos.
2. **View Feed**: Users can scroll through a timeline of posts from friends/followed users.
3. **Like & View Counter**: Users can like posts and view real-time count of likes and views.

### Non-Functional Requirements (How well it performs)
1. **Ultra-Low Latency**: Feed should load within 200–500ms, even with an empty browser cache.
2. **High Availability**: If one server crashes, the site shouldn't go down (99.99% uptime).
3. **Scalability**: Handle 500 million daily active users (DAU) and high-concurrency traffic.
4. **Eventual Consistency**: It's okay if a friend sees your like 1–2 seconds late, as long as the system stays fast and doesn't crash.

---

### 2.1 Back-of-the-Envelope Estimation (Traffic & Storage Math)

In system design interviews, estimating scale helps determine server capacity, database size, and caching requirements.

#### A. User & Activity Assumptions
- **Monthly Active Users (MAU)**: **1 Billion**
- **Daily Active Users (DAU)**: **500 Million** (50% daily active engagement rate)
- **Time in a day**: $86,400 \text{ seconds} \approx \mathbf{100,000 \text{ seconds}}$ (rounded up for easy mental math).

#### B. Traffic & QPS (Queries Per Second)

1. **Feed Read Requests (Fetch Feed)**:
   - Assume each DAU checks their feed **5 times a day**.
   - Total Feed Reads per day = $500 \text{M} \times 5 = \mathbf{2.5 \text{ Billion reads/day}}$.
   - **Average Read QPS** = $\frac{2.5 \text{ Billion}}{100,000 \text{ seconds}} = \mathbf{25,000 \text{ QPS}}$.
   - **Peak Read QPS** ($2\times \text{Average}$) = $\mathbf{50,000 \text{ QPS}}$.

2. **Post Write Requests (Creating Posts)**:
   - Assume users post an average of **0.2 posts a day** (1 post every 5 days per user).
   - Total Posts per day = $500 \text{M} \times 0.2 = \mathbf{100 \text{ Million posts/day}}$.
   - **Average Write QPS** = $\frac{100 \text{ Million}}{100,000 \text{ seconds}} = \mathbf{1,000 \text{ QPS}}$.
   - **Peak Write QPS** ($2\times \text{Average}$) = $\mathbf{2,000 \text{ QPS}}$.

3. **Likes & Reactions**:
   - Assume users like **10 posts a day**.
   - Total Likes per day = $500 \text{M} \times 10 = \mathbf{5 \text{ Billion likes/day}}$.
   - **Average Like QPS** = $\frac{5 \text{ Billion}}{100,000 \text{ seconds}} = \mathbf{50,000 \text{ QPS}}$.
   - **Peak Like QPS** = $\mathbf{100,000 \text{ QPS}}$.

---

#### C. Storage Estimations (Database & Media)

1. **Post Text & Metadata Storage**:
   - Each post payload (Post ID, User ID, Text content, Timestamp, Media URLs list) $\approx \mathbf{1 \text{ KB}}$.
   - Daily Post Metadata Storage = $100 \text{ Million posts} \times 1 \text{ KB} = \mathbf{100 \text{ GB / day}}$.
   - **5-Year Storage** = $100 \text{ GB/day} \times 365 \text{ days} \times 5 = \mathbf{182.5 \text{ TB}}$.

2. **Media Storage (Images & Videos)**:
   - Assume 20% of posts contain media (20 Million media posts/day).
   - Average media size (compressed image or short video clip) $\approx \mathbf{500 \text{ KB}}$.
   - Daily Media Storage = $20 \text{ Million} \times 500 \text{ KB} = \mathbf{10 \text{ TB / day}}$.
   - **5-Year Media Storage** = $10 \text{ TB/day} \times 365 \times 5 = \mathbf{18.25 \text{ Petabytes (PB)}}$ (Stored in S3 / Blob storage).

---

#### D. RAM / Memory Estimations (Redis Feed Cache)

- We store the top **50 post IDs** in Redis for each active user's pre-computed feed.
- Each post ID + score $\approx 20 \text{ bytes}$.
- $50 \text{ post IDs} \times 20 \text{ bytes} = 1 \text{ KB}$ per user feed cache.
- RAM required for 500M DAU = $500 \text{ Million users} \times 1 \text{ KB} = \mathbf{500 \text{ GB of RAM}}$.
- With Redis cluster replication factor of 3 (Primary + 2 Replicas), total RAM required = $500 \text{ GB} \times 3 = \mathbf{1.5 \text{ TB of RAM}}$.

---

## 3. Architecture: Monolith vs. Microservices

### What is a Monolith?
A single large application where user management, post creation, feed building, and likes are all bundled into one codebase and deployed together.

### What are Microservices?
Breaking the system into smaller, independent applications (services) that communicate over the network. Each service manages its own responsibility and data storage.

### Why Microservices for a News Feed?
| Feature | Monolith | Microservices (Chosen) |
| :--- | :--- | :--- |
| **Scaling** | Must scale the *entire* app even if only "Likes" need more resources. | Scale individual services (e.g., scale the Feed Service 10x while keeping User Service small). |
| **Fault Isolation** | If the Video Processing module crashes, the whole app dies. | If Video Processing crashes, users can still view text feeds and like posts. |
| **Team Velocity** | Multiple teams work on the same codebase, creating merge conflicts. | Teams independently build, test, and deploy their own services. |

---

## 4. Service Boundaries & Responsibilities

Here is how we split our system into specialized microservices:

```
[ User Client (Mobile / Web) ]
               │
               ▼
       [ API Gateway ]  ───► (Authentication, Rate Limiting, Routing)
               │
   ┌───────────┼───────────────────┬───────────────────┐
   ▼           ▼                   ▼                   ▼
[User Service] [Post Service]   [Feed Service]   [Counter Service]
   │           │                   │                   │
   ▼           ▼                   ▼                   ▼
(User DB)   (Post DB & Media)   (Feed Cache)      (Redis Counters)
```

1. **API Gateway**: The entry point. Handles authentication, SSL termination, and rate limiting (preventing spam/DDoS).
2. **User Service**: Manages user profiles, friend lists, and follow/unfollow relationships.
3. **Post Service**: Handles creating new posts, uploading text/metadata, and triggering media processing.
4. **Media Processing Service**: Compresses images and transcodes uploaded videos into multiple resolutions for fast streaming.
5. **Feed Service**: Generates, stores, and serves personalized timelines for every user.
6. **Counter / Reaction Service**: Tracks likes, comments count, and view counts with high throughput.

---

## 5. API Design (High Level)

APIs define how the client (mobile app or browser) talks to our backend services.

### 1. Create a Post
- **Endpoint**: `POST /v1/posts`
- **Payload**:
  - `user_id`: ID of the post author
  - `content`: Text content
  - `media_urls`: Array of uploaded image/video URLs
- **Response**: `201 Created` with `post_id` and timestamp.

### 2. Get User Feed
- **Endpoint**: `GET /v1/feed`
- **Query Parameters**:
  - `page_token` or `cursor`: For pagination (infinite scrolling)
  - `limit`: Number of posts to fetch (e.g., 10 or 20)
- **Response**: List of post objects containing author info, text, media links, like count, and view count.

### 3. Like a Post
- **Endpoint**: `POST /v1/posts/{post_id}/like`
- **Response**: `200 OK` (Acknowledged asynchronously).

---

## 6. Deep Dive 1: Feed Generation (Push vs. Pull vs. Hybrid)

How do we build a user's feed? There are 3 main approaches:

---

### Strategy A: Fan-out on Read (Pull Model)

**How it works**:
- When you post something, it is stored only in your own post list.
- When your friend logs in and opens their feed, the system looks up all 500 of their friends, pulls the latest posts from each friend, blends them together, sorts them by timestamp, and displays them.

```
User A Posts ──► Saved in User A's DB table only.

User B opens app ──► Fetch User B's friend list (500 friends)
                 ──► Query DB for latest posts of ALL 500 friends
                 ──► Merge & Sort in memory ──► Return to User B
```

**Pros**:
- **Fast writes**: Posting is instantaneous because you write to only one place.
- **No wasted storage**: Posts are not duplicated into thousands of feeds.

**Cons**:
- **Slow reads**: Fetching and merging posts from hundreds of friends on every single feed request causes massive database load and slow response times.

---

### Strategy B: Fan-out on Write (Push Model)

**How it works**:
- Every user has a pre-computed "Inbox" or "Feed Cache" stored in memory (e.g., Redis).
- When a user posts something, a background worker **pushes** that post ID into the feed cache of **every single follower/friend**.
- When you open your app, the Feed Service simply reads your pre-computed Redis inbox in milliseconds!

```
User A Posts ──► Background Worker finds User A's 500 friends
             ──► Writes Post ID directly into 500 friends' Redis Feed Caches

User B opens app ──► Read User B's Redis Feed Cache directly (Instant!)
```

**Pros**:
- **Ultra-fast reads**: Feed loads instantly because it's already pre-built in memory!
- **Low database stress during reads**.

**Cons**:
- **The Celebrity Problem**: If a user with 50 million followers posts, the system must perform 50 million write operations immediately! This causes system lag, queues backing up, and storage bloat.

---

### Strategy C: The Hybrid Model (Best of Both Worlds)

To solve the limitations of both models, modern systems use a **Hybrid Approach**:

1. **For Normal Users (e.g., < 10,000 followers)**: Use **Push Model (Fan-out on Write)**. Pre-build feeds in Redis caches for instant loading.
2. **For Celebrities / High-Follower Users (e.g., > 10,000 followers)**: Use **Pull Model (Fan-out on Read)**.
   - Do **NOT** push a celebrity's post to millions of followers' feed caches.
   - Instead, when a follower opens their feed:
     - Fetch their pre-computed Push Feed (from normal friends).
     - Fetch the latest posts from the celebrities they follow.
     - Merge the two sets in memory before returning to the user.

```
                                 ┌── [ Normal Friend ] ──► PUSH to User's Feed Cache
[ User opens Feed ] ──► Blends ──┤
                                 └── [ Celebrity Post ] ──► PULL on-demand & merge
```

---

## 7. Database Strategy (SQL vs. NoSQL)

No single database fits every requirement. We use a **polyglot persistence** approach (choosing the right database for the right job).

```
                      ┌──► Relational DB (PostgreSQL / MySQL) ──► User Profiles, Friend Graph
                      │
[ Data Storage Layer ]├──► NoSQL Document DB (Cassandra / DynamoDB) ──► Posts Metadata
                      │
                      ├──► In-Memory Cache (Redis) ──► Timeline Feed Caches & Counters
                      │
                      └──► Object Storage (AWS S3 / Blob) ──► Images & Videos
```

### 1. User & Graph Data (Relationships)
- **Database**: Relational DB (MySQL / PostgreSQL) or Graph DB (Neo4j).
- **Why**: User profiles and friend connections require strong consistency, structural constraints, and relational queries (e.g., "Find mutual friends").

### 2. Posts Data
- **Database**: Distributed NoSQL Key-Value / Wide-Column Store (Apache Cassandra or AWS DynamoDB).
- **Why**:
  - High write throughput.
  - Horizontally scalable across multiple database nodes.
  - Simple access pattern: Read post by `post_id` or query posts by `user_id` ordered by time.

### 3. Media Storage (Images & Videos)
- **Database**: Blob / Object Storage (AWS S3, Google Cloud Storage).
- **Why**: Databases are not optimized for storing large binary files like HD photos or MP4 videos.

### Database Sharding Strategy
As data grows beyond what one database server can hold, we **shard** (partition) the database across multiple machines:
- **Shard Key for Posts**: Partition posts by `user_id` using consistent hashing. All posts from a user live on the same shard, making user queries fast.

---

## 8. Deep Dive 2: Real-Time Likes & Views (High Concurrency)

### The "Like Problem" Explained

Why is liking a post so difficult at Facebook scale?

1. **Row Locking & DB Crashes**:
   - When a post goes viral (e.g., a World Cup post), **100,000 users click "Like" in the exact same second**.
   - If you execute a direct SQL query: `UPDATE posts SET like_count = like_count + 1 WHERE post_id = 999;`
   - All 100,000 requests try to update the **exact same row** in the database.
   - The database locks row `999`. 99,999 other connections wait in queue, running out of connection pools, spiking CPU to 100%, and crashing the database immediately.

2. **The Deduplication Challenge**:
   - A user can only like a post **once**.
   - We must check: *"Has User 123 already liked Post 999?"*
   - Querying `SELECT * FROM post_likes WHERE user_id = 123 AND post_id = 999;` 100,000 times/second disk-reads will destroy database performance.

3. **Latency vs. Consistency Trade-off**:
   - The user who clicks "Like" expects their heart icon to turn red **instantly** (0ms latency).
   - But other users across the globe don't care if the like count updates from `1,000,000` to `1,000,001` with a 1–2 second delay. **Eventual consistency is key**.

---

### Step-by-Step Architecture Solution Flow

Here is the exact end-to-end journey of a "Like" request through our high-scale system:

```
[ 1. User Clicks Like ] ──► Mobile App turns Heart RED instantly (Local Optimistic UI update)
           │
           ▼
[ 2. API Gateway ]      ──► Rate limiting check ──► Forward to Counter Service
           │
           ▼
[ 3. Redis Deduplication Check ] ──► SADD post:999:liked_by user:123
           │                       (If user already exists in Set ──► Stop & return success)
           ▼
[ 4. Sharded Redis Counter ]     ──► Pick random shard (e.g., post:999:likes:shard_4)
           │                       INCRBY 1 (Ultra-fast in-memory increment)
           ▼
[ 5. Message Queue (Kafka) ]     ──► Publish PostLikedEvent(post_id=999, user_id=123)
           │                       (Buffers write traffic to protect DB)
           ▼
[ 6. Worker Service ]            ──► Accumulates events for 5 seconds (e.g. 5,000 likes)
           │                       Flushes 1 Single Batch Write to DB:
           ▼                       "UPDATE posts SET like_count = like_count + 5000 WHERE id = 999;"
[ 7. Main Database ]             ──► Persistent storage updated cleanly with ZERO row locking!
```

---

### Step-by-Step Walkthrough

#### Step 1: Optimistic Local UI Update (Client-Side)
- When the user taps the heart icon, the mobile app **immediately** turns the icon red without waiting for the server response.
- If the network request fails later, the app reverts the icon and shows a subtle error toast.

#### Step 2: Rate Limiting & Gateway Routing
- The request hits the API Gateway (`POST /v1/posts/999/like`).
- Rate limiting prevents bot scripts from sending thousands of like requests per second from a single IP.

#### Step 3: Fast Deduplication via Redis Sets & Bucket Sharding
- **How Redis Deduplicates Natively**:
  We use a **Redis Set** data structure with the `SADD` (Set Add) command:
  `SADD post:999:liked_by user:123`
  - In a Set, every item must be unique.
  - If `user:123` is **not** in the set, Redis adds `user:123` and returns `1` (New Like).
  - If `user:123` is **already** in the set, Redis does nothing and returns `0` (Duplicate Like).

- **How Sharding Solves the Bottleneck for Viral Posts**:
  - If 1 million users like the *same* post (`post:999`), a single `post:999:liked_by` Redis set would live on 1 node, bottlenecking that single node's CPU.
  - **Solution**: We shard the deduplication set into `10 buckets` using `hash(user_id) % 10`:
    - `post:999:liked_by:bucket_0` (on Node 1)
    - `post:999:liked_by:bucket_1` (on Node 2)
    - ...
    - `post:999:liked_by:bucket_9` (on Node 10)
  - Because `user:123` **always hashes to the exact same bucket** (e.g. Bucket 4), all duplicate clicks from `user:123` hit Bucket 4 on Node 5 and get correctly rejected with `0`.
  - Meanwhile, requests from 1 million users are spread evenly across all 10 Redis nodes!

#### Step 4: Sharded In-Memory Increment
- To prevent a single Redis node from becoming a bottleneck for a viral post, we split the counter into **10 shards** (`post:999:likes:shard_1` through `shard_10`).
- The server randomly picks one shard (e.g. Shard 4) and increments it in memory: `INCRBY post:999:likes:shard_4 1`.
- Memory operations take less than **1 millisecond**.

#### Step 5: Event Buffering via Message Queue (Kafka)
- The Counter Service publishes a `PostLikedEvent` to **Apache Kafka**.
- Kafka acts as a **buffer / shock absorber**. Even if 200,000 likes arrive in 1 second, Kafka queues them safely without dropping data or overwhelming downstream systems.

#### Step 6: Batch Aggregation by Background Workers
- A pool of Worker Services listens to Kafka.
- Instead of writing every single like to the database, workers **batch-aggregate** incoming events in memory over 3–5 second windows.
- *Example*: In 5 seconds, 5,000 likes arrive for Post 999. The worker aggregates this into a single instruction: `+5000 likes for Post 999`.

#### Step 7: Write-Back Batch Update to Database
- The worker executes **1 single database query**:
  `UPDATE posts SET likes_count = likes_count + 5000 WHERE id = 999;`
- Plus 1 bulk insert into the audit table:
  `INSERT INTO post_likes (post_id, user_id) VALUES (999, 123), (999, 124), ... [5,000 rows];`
- The database performs 1 fast update instead of 5,000 row-locking operations!

---

## 9. Deep Dive 3: Caching Strategy & Fast Feed Load

To achieve instant feed loading (even with an empty browser cache), caching must exist at multiple layers:

```
[ Client App ]
     │
     ▼
[ Content Delivery Network (CDN) ] ──► Caches Images & Videos near user geographically
     │
     ▼
[ API Gateway ]
     │
     ▼
[ Application Layer ]
     │
     ▼
[ Distributed Cache (Redis Cluster) ] ──► Caches User Feed Post IDs & Post Content Metadata
     │
     ▼
[ Main Database ]
```

### 1. Content Delivery Network (CDN) for Media
- All images and videos are stored in S3 and cached on CDN servers globally (e.g., Cloudflare, CloudFront).
- When a user in Tokyo requests an image, it is served from a local Tokyo CDN edge server instead of fetching it from a US database server.

### 2. Redis Caching for Feed Timelines

To render a feed instantly, every active user has a pre-computed timeline stored in Redis as a **Sorted Set (`ZSET`)**.

#### Why Redis Sorted Set (`ZSET`)?
In a Redis `ZSET`, every element has two parts:
- **Member / Value**: The `post_id` (e.g. `post_5001`).
- **Score**: The Unix timestamp when the post was created (e.g. `1700000100`).

Redis automatically keeps items in a `ZSET` ordered by score. This allows us to fetch the top 10 newest posts in sub-millisecond time using `ZREVRANGE`.

---

#### Concrete Step-by-Step Example

Let me walk you through an example with **Alice** (`user_101`), who follows **Bob** (`user_202`) and **Charlie** (`user_303`).

```
[ Alice's Redis Key ]: "feed:user_101"
 ┌─────────────────────────────────────────────────────────────┐
 │ SCORE (Timestamp) │ MEMBER (Post ID)                        │
 ├───────────────────┼─────────────────────────────────────────┤
 │ 1700000200        │ post_5002  (Posted by Charlie at 10:03)  │
 │ 1700000100        │ post_5001  (Posted by Bob at 10:01)      │
 │ 1700000000        │ post_4999  (Older post at 09:55)        │
 └─────────────────────────────────────────────────────────────┘
```

1. **Step 1: Bob publishes a new post** (`post_5001` at 10:01 AM):
   - The system executes a push command into Alice's Redis feed:
     `ZADD feed:user_101 1700000100 post_5001`

2. **Step 2: Charlie publishes a new post** (`post_5002` at 10:03 AM):
   - The system pushes Charlie's post into Alice's feed:
     `ZADD feed:user_101 1700000200 post_5002`

3. **Step 3: Alice opens her mobile app (Fetching Feed)**:
   - Alice's app requests her feed: `GET /v1/feed?limit=10`
   - The Feed Service queries Redis for the 10 newest posts:
     `ZREVRANGE feed:user_101 0 9`
   - Redis instantly returns: `["post_5002", "post_5001", "post_4999"]` (in < 1ms!).

4. **Step 4: Hydration (Fetching Post Details)**:
   - Now we have the post IDs, but we need the actual post text, image links, author name, and like counts.
   - The server performs a bulk read from the Redis Key-Value cache:
     `MGET post:post_5002 post:post_5001 post:post_4999`
   - The complete feed data is sent back to Alice's phone!

5. **Step 5: Memory Trimming**:
   - To keep Redis RAM low, we don't store 10,000 posts in Alice's cache.
   - We trim the set to keep only the 500 newest posts:
     `ZREMRANGEBYRANK feed:user_101 0 -501` (Deletes older posts beyond index 500).

---

### 3. Pre-Fetching & Warm Cache
- **Cold Start**: When an inactive user logs in after weeks, their cache is empty ("cold").
- **Warm Cache Strategy**: 
  - Predict when users usually log in (e.g., morning/evening).
  - Pre-generate their feeds before they open the app.
  - Cache the top 50 posts so the initial view renders in under 100ms.

### 4. Cache Invalidation
- Use Time-To-Live (**TTL**) on cached items (e.g., expire feed cache entries after 7 days).
- Use **LRU (Least Recently Used)** eviction policy so active users stay in memory while inactive users' feed memory is freed up.

### 5. Redis Sharding & Redis Cluster Strategy
A single Redis instance maxes out around 100,000 requests/sec and ~256GB of RAM. At Facebook scale (millions of concurrent users), **a single Redis instance will instantly crash**. Therefore, Redis is **sharded into a cluster of nodes**:

- **Sharding by User ID (Feed Caches)**:
  - User feeds (`feed:user_id`) are distributed across different Redis nodes using **Consistent Hashing**.
  - `feed:user_123` maps to Node A, `feed:user_456` maps to Node B.
- **Sharding by Post ID (Post Metadata & Counters)**:
  - Post metadata and reaction keys (`post:post_id:likes`) are hashed across nodes.
- **Counter Sharding (Hotspot Prevention)**:
  - For viral posts, a single key on Node A would still get overwhelmed. We split the single key into `N` sub-keys (`post_id:likes_shard_1` through `shard_10`) spread across multiple Redis nodes in the cluster.

---

## 10. Solving the Celebrity / Viral Content Problem

When a post receives millions of likes and views per minute, it creates a **Hotspot Key** problem in caches and databases. Here is how we mitigate it:

### 1. Counter Sharding (Distributed Counters)
Instead of storing a celebrity post's like counter in a single Redis key (`post:999:likes`), split the counter into **N separate shards**:
- `post:999:likes:shard_1`
- `post:999:likes:shard_2`
- `...`
- `post:999:likes:shard_10`

When a user likes the post, randomly pick one of the 10 shards to increment. To display the total count, sum up all 10 shards. This distributes traffic across 10 memory nodes!

### 2. Local In-Memory Caching (L1 Cache)
For viral post text/metadata, cache the post directly in the memory of the API servers themselves (L1 cache) for 1–2 seconds. This shields even Redis from being overwhelmed by millions of requests for the exact same post.

### 3. Asynchronous View Counter Aggregation
View counts do not need to be 100% accurate down to the exact single view.
- Collect view counts on the client side (e.g., aggregate 20 post views into 1 batch API call).
- Flush batches asynchronously to Kafka, processing counters in bulk.

---

## 11. Summary Checklist for Interviews

When asked this question in an interview, present your solution in this logical sequence:

1. **Clarify Requirements**: State functional (posts, feed, likes) and non-functional goals (latency, scale, availability).
2. **High-Level Microservices Architecture**: API Gateway, User Service, Post Service, Feed Service, Counter Service.
3. **Feed Generation Strategy**: Explain Push vs. Pull, and pivot to the **Hybrid Model** to handle celebrities.
4. **Database Selection**: Relational for Graph/User Data, Cassandra/NoSQL for Posts, Redis for Feed Caches, S3/CDN for Media.
5. **High Concurrency & Counters**: Explain Redis + Kafka async queue + batch database updates for Likes/Views.
6. **Caching & Latency**: CDN for media, Redis ZSET for timeline feeds, warm cache for fast login loading.
7. **Viral / Edge Cases**: Counter sharding and API-level L1 caching for hotspot posts.

---
