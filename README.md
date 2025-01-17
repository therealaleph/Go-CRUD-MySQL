<h1>Go CRUD MySQL</h1>

This repository contains a Go-based REST API that demonstrates asynchronous handling of database operations for a MySQL database. The API provides endpoints for creating, reading, updating, and deleting data in a specified table.

<h2>Features:</h2>
Asynchronous operations: Uses goroutines and db.ExecContext to perform database operations asynchronously, improving scalability and responsiveness. <br>
Authentication: Requires a Bearer token for authorization to access the API endpoints. <br>
Flexible data handling: Accepts and returns JSON data for database operations. <br>
Modular design: The API is structured into modules for easy maintenance and extensibility. <br>
