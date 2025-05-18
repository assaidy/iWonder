### 🔍 **iWonder** 

**iWonder** is a question-and-answer app inspired by Stack Overflow.  

#### ✨ **Features**  
- [x] **Ask & Answer**: Post questions and contribute answers.
- [x] **Voting System**: Upvote/downvote to highlight the best responses.
- [x] **Solutions**: Mark a specific comment as a solution for your question.
- [x] **Tags**: Organize content by topics (e.g., tech, lifehacks).
- [ ] **Real-time Notifications**: Stay updated on responses and mentions.

#### 🛠️ **Tech Stack**  
- Fiber: Web framework for building the API.
- JWT (JSON Web Tokens): For user authentication.
- PostgreSQL: Relational database for data storage.
- Goose: Database migration tool for managing schema changes.
- SQLC: SQL compiler and type-safe query generator for Go.
- Docker Compose: Containerization and orchestration for local development and deployment,
- Cursor Pagination: Implemented cursor-based pagination to minimize bandwidth usage and reduce server load, ensuring efficient data retrieval for large datasets.
- Prefork: Enabled for load balancing and improved performance under high traffic.
