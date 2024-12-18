# Library Management System

This project is a simple Library Management API built with **Go**, **PostgreSQL**, and Docker Compose. It allows you to manage users, books, borrowing, and returning books.

---

## **1. Prerequisites**

To use the project, the following tools are necessary

- [Docker](https://www.docker.com/)
- [Docker Compose](https://docs.docker.com/compose/)
- [Go](https://golang.org/dl/)

---

## **2. Project Structure**

The project directory looks like this:

```
.
├── .env                # Environment variables (needs to be added because manually)
├── docker-compose.yml  # Docker Compose file to run containers
├── Dockerfile          # Dockerfile for building the Go application
├── go.mod              # Go dependencies
├── go.sum              # Go checksum file
└── main.go             # Go application entry point
```

---

## **3. Environment Variables**

The database credentials are stored in a `.env` file. Create a `.env` file in the root directory and for example add the following:

```plaintext
DB_USER=postgres
DB_PASSWORD=12345678
DB_HOST=db
DB_PORT=5432
DB_NAME=library
```

---

## **4. Running the Application**

To run the application and the PostgreSQL database using Docker Compose:

### Step 1: Build and Start Containers
Run the following command in the terminal:

```bash
docker-compose up --build
```

This will:
1. Start the PostgreSQL database.
2. Start the Go application on port `8080`.

---

## **5. Testing the Application**

### API Endpoints:

| Method | Endpoint                 | Description                 |
|--------|--------------------------|-----------------------------|
| GET    | `/displayBooks`          | List all books              |
| POST   | `/addUser`               | Add a new user              |
| GET    | `/displayUsers`          | List all users              |
| POST   | `/borrowBook`            | Borrow a book               |
| POST   | `/returnBook`            | Return a book               |

---

### Example Request for GET: Display Users

1. **Method**: `GET`  
2. **URL**: `http://localhost:8080/displayUsers`  
3. **Response**:
   ```json
   [
       {
           "id": 1,
           "first_name": "John",
           "last_name": "Doe"
       },
       {
           "id": 2,
           "first_name": "Jane",
           "last_name": "Smith"
       }
   ]
   ```

---

### Example Request for POST: Add a User

1. **Method**: `POST`  
2. **URL**: `http://localhost:8080/addUser`  
3. **Body** (JSON):
   ```json
   {
       "first_name": "John",
       "last_name": "Doe"
   }
   ```

4. **Response**:
   ```json
   {
       "id": 1,
       "first_name": "John",
       "last_name": "Doe"
   }
   ```

---

## **6. Accessing the Database**

To access the PostgreSQL database manually:

1. Connect to the `db` container:
   ```bash
   docker exec -it library_database psql -U postgres -d library
   ```

2. Run queries:
   ```sql
   SELECT * FROM books;
   ```

---

## **7. Stopping the Application**

To stop all containers:

```bash
docker-compose down
```

---

---

