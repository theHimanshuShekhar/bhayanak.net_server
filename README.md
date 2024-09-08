# BChat Backend Server

BChat is an IRC-like chat application that supports anonymous communication in different global chatrooms. The backend server is built using **WebSockets** and provides the real-time messaging infrastructure for both the terminal-based and web-based clients. Users can join chatrooms, exchange messages, and enjoy a seamless, low-latency chat experience.

## Features

- **WebSocket-based Real-time Communication**: Provides real-time, full-duplex communication between the server and clients.
- **Anonymous Chat**: No registration or login required; users can chat anonymously.
- **Global Chatrooms**: Users can join and chat in different global chatrooms.
- **Support for Multiple Clients**:
  - Terminal-based client (CLI)
  - Web-based client
- **Scalable Architecture**: Designed to handle multiple connections efficiently.
- **Secure Communication**: Basic token-based authentication for secure WebSocket connections (can be extended to use JWT or OAuth).

## Table of Contents

- [Project Structure](#project-structure)
- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Installation](#installation)
  - [Running the Server](#running-the-server)
- [Configuration](#configuration)
- [Authentication](#authentication)
- [Contributing](#contributing)
- [License](#license)

## Project Structure

The project is structured as follows:

```
├───cmd
│ └───server
├───internal
│ ├───server
│ │ ├───handler.go
│ │ └───server.go
│ ├───config
│ │ └───config.go
│ └───services
│   └───auth.go
│   └───chatroom.go
│   └───message.go
│   └───user.go
│   └───database.go
└───pkg
│    ├───logger
│    │ └───logger.go
│    └───utils
│      └───utils.go
├───go.mod
├───Makefile
└── README.md
```

## Getting Started

### Prerequisites

Ensure that you have the following installed:

- [Go](https://golang.org/doc/install) (version 1.18 or higher)
- [Git](https://git-scm.com/)
- (Optional) [Docker](https://www.docker.com/) for containerized deployment

### Installation

1. **Clone the repository:**

   ```bash
   git clone https://github.com/theHimanshuShekhar/bchat-server.git
   cd bchat-server
   ```

2. Install the dependencies:

   ```bash
   go mod tidy
   ```

### Running the Server

1. Setup the environment variables:
   the server reads configuration from environment variables, such as the port number and secret key. You can set these variables in a `.env` file in the root directory of the project. Here's an example:

   ```bash
   PORT=8080
   SECRET_KEY=secret
   ```

2. Start the server:

   ```bash
   go run cmd/server/main.go
   ```

   or

   ```bash
   make run # if you have make installed
   ```

3. Connect a client to the server:

   You can use either a terminal-based client or the web-based client to connect to the server and start chatting in global chatrooms.

## Configuration

The server reads configuration from environment variables. Here's a list of the available configuration options:

| Environment Variable | Description                                                               | Default Value |
| -------------------- | ------------------------------------------------------------------------- | ------------- |
| `PORT`               | The port number on which the server will listen for incoming connections. | `8080`        |
| `SECRET_KEY`         | The secret key used for token-based authentication.                       | `secret`      |

## Authentication

The server uses token-based authentication for secure WebSocket connections. The token is generated using a secret key and a user ID. The token is sent in the `Authorization` header of the WebSocket connection request. The server verifies the token and allows the connection if it is valid.

## Contributing

Contributions are welcome! If you find a bug or have a suggestion, please open an issue or submit a pull request on the GitHub repository.

## License

This project is licensed under the GNU AGPLv3 License. See the [LICENSE](LICENSE) file for more information.
