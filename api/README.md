# Code Roast API

This is the backend API for the "Code Roast" forum/community, built using Go and a variety of modern technologies.  It provides the core functionality for users to share jokes, roast programming languages (in a lighthearted way), discuss new technologies, and have serious discussions about programming topics.

## Technologies Used

*   **Backend:** Go
*   **Database:** PostgreSQL
*   **Caching:** Redis
*   **Authentication:** JWT (JSON Web Tokens)
*   **Real-time Updates (Optional):** WebSockets (using `gorilla/websocket`)
*   **CI/CD:** GitHub Actions
*   **Deployment:** Docker, Docker Compose
*   **Logging:** `logrus` (or `zap`) structured logging library

## Features

*   **User Authentication:** Users can register, log in, and manage their profiles. JWT is used for secure authentication.
*   **Categories/Subforums:** The forum is organized into categories (e.g., "Language Roasts," "Framework Fails," "New Tech Discussions," "Serious Help") to help users find relevant content.
*   **Posts/Threads:** Users can create new posts in the appropriate category. Threading (replies to replies) is implemented for discussions.
*   **Voting/Reactions:** A voting system (upvotes/downvotes) or reactions (similar to Facebook's reactions) is implemented for posts and replies.
*   **Image/GIF Support:** Users can upload images and GIFs in their posts and replies.
*   **Code Snippet Highlighting:** Code highlighting is supported for sharing code snippets, especially in the "Serious Help" category.
*   **Search:** Users can search for posts based on keywords.
*   **Real-time Updates (Optional):** Real-time updates for new posts/replies within a thread are available via WebSockets.
*   **Admin Features (Optional):**  Admin features like deleting posts/threads, managing users, or moderating content are available.

## Getting Started

### Prerequisites

*   Go (latest version recommended)
*   PostgreSQL
*   Redis
*   Docker
*   Docker Compose

### Installation

1.  Clone the repository:

    ```bash
    git clone [https://github.com/your-username/code_roast_api.git](https://www.google.com/search?q=https://github.com/your-username/code_roast_api.git)  # Replace with your repo URL
    cd code_roast_api
    ```

2.  Set up environment variables:

    Create a `.env` file in the root directory and add the following environment variables:

    ```
    DATABASE_URL=postgres://user:password@host:port/database_name
    REDIS_URL=redis://host:port
    JWT_SECRET=your_jwt_secret_key  # Generate a strong secret key
    # ... other environment variables
    ```

3.  Run the application using Docker Compose:

    ```bash
    docker-compose up -d --build
    ```

### API Endpoints

(Document all your API endpoints here with details about request methods, parameters, request bodies, and response formats.  Use a tool like Swagger or OpenAPI for more comprehensive API documentation.)

**Example:**

*   `POST /api/users/register`: Registers a new user.

    *   Request Body:

        ```json
        {
            "username": "john_doe",
            "email": "[email address removed]",
            "password": "password123",
            "fullname": "John Doe"
        }
        ```

    *   Response:

        ```json
        {
            "message": "User registered successfully"
        }
        ```

*   `GET /api/posts`: Retrieves a list of posts.

    *   Query Parameters:
        *   `category`: (Optional) Filter posts by category.
        *   `page`: (Optional) Page number for pagination.
        *   `limit`: (Optional) Number of posts per page.

    *   Response:

        ```json
        [
            // Array of post objects
        ]
        ```

(Add documentation for all other endpoints.)

## Running Tests

```bash
go test ./...  # Run all tests
```