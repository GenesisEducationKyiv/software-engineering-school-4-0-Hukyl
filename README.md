# genesis-kma-school-entry
Test task for Genesis &amp; KMA Software Engineering School 4.0

## Installation Steps

### Local Development

1. Install Golang by following the official installation guide: [Golang Installation Guide](https://golang.org/doc/install)

2. Clone the repository:
    ```bash
    git clone https://github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/currency-rate.git
    ```

3. Navigate to the project directory:
    ```bash
    cd genesis-kma-school-entry
    ```

4. Copy the `.env.sample` file and rename it to `.env` (make sure `.env` file is on the same directory level as `docker-compose.yml`):
    ```bash
    cp .env.sample .env
    ```

5. Install the project dependencies:
    ```bash
    go mod download
    ```

6. Start the local development server:
    ```bash
    go run main.go
    ```

The base url will be available at `localhost:<port>`.

### Running Services in Docker Compose

1. Install Docker by following the official installation guide: [Docker Installation Guide](https://docs.docker.com/get-docker/)

2. Clone the repository:
    ```bash
    git clone https://github.com/your-username/genesis-kma-school-entry.git
    ```

3. Navigate to the project directory:
    ```bash
    cd genesis-kma-school-entry
    ```

4. Copy the `.env.sample` file and rename it to `.env` (make sure `.env` file is on the same directory level as `docker-compose.yml`):
    ```bash
    cp settings/.env.sample .env
    ```
    *NOTE*: some of the variables are hardcoded into `docker-compose.yml` to ensure that the API uses the container database. 

5. Build and start the Docker containers:
    ```bash
    docker compose up -d
    ```

The base url will be available at `127.0.0.1:<port>`.

## Endpoints

The endpoints fully conform to the provided endpoint schemas in Swagger documentation.

### Get USD-UAH rate

- Method: `GET`
- URL: `/rate`
- Purpose: provides a USD-UAH current rate.

### Subscribe to email notifications

- Method: `POST`
- URL: `/subscribe`
- Form-data parameter: `email`
- Purpose: subscribe to daily email notifications of rate

## Testing

Most of the subpackages are covered by unittests.
In order to run tests:

```bash
go test -v ./...
```

## License
This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Copyright
© 2024-current Andrii Shalaiev. All Rights Reserved.
