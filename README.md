# Real-Time RPC Chat System -- Go Concurrency Implementation

This project is an enhanced version of an RPC-based multi-client chat
system.\
The original version returned the full chat history whenever a client
sent a message.\
The updated system now uses **real-time broadcasting** implemented
through **Go concurrency**, providing instant message delivery across
all connected clients.

------------------------------------------------------------------------

## 📌 Features

### ✔ Real-Time Broadcasting

-   When a client joins, the server notifies all clients:\
    **"User \[ID\] joined"**
-   When any client sends a message, the server broadcasts it to every
    other client.
-   No self-echo --- the sender does not receive their own message.

### ✔ Concurrency with Goroutines & Channels

-   The server runs a broadcast goroutine to handle all real-time
    messages.
-   Each client has its own receiving goroutine and message channel.
-   Non-blocking communication ensures smooth handling of multiple
    clients.

### ✔ Thread-Safe Shared State

-   A `sync.Mutex` protects the connected clients list.
-   Safe addition/removal of clients in concurrent scenarios.

### ✔ RPC for Sending / Registration Only

-   RPC calls:
    -   `RegisterClient` → connect a new client
    -   `SendMessage` → send messages to the server
-   Receiving messages is handled through goroutines, not RPC returns.

------------------------------------------------------------------------

## 🧩 Architecture Overview

    SERVER
     ├── Handles RPC calls
     ├── Manages client map (Mutex)
     ├── Broadcast goroutine
     └── Channels to every client

    CLIENT
     ├── RPC sender goroutine
     └── Receiving goroutine (listens to server channel)

------------------------------------------------------------------------

## ⚙️ Running the Project

### Start the Server

    go run server.go

### Start Clients

    go run client.go

You can open multiple terminals and run multiple clients simultaneously.

------------------------------------------------------------------------

## 📁 Repository Structure

    .
    ├── server.go
    ├── client.go
    └── README.md

------------------------------------------------------------------------

# 📘 Quick Conclusion

This task successfully transforms a basic RPC chat system into a **fully
concurrent, real-time broadcast chat application** using Go. By
combining goroutines, channels, and mutex-protected shared state, the
system now supports instant communication between multiple clients
without blocking or message duplication. The design demonstrates
practical use of Go's concurrency model and shows how RPC can be used
alongside channel-based asynchronous message delivery to build
responsive distributed applications.
