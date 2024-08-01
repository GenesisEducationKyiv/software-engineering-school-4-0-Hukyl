# genesis-kma-school-entry
Test task for Genesis &amp; KMA Software Engineering School 4.0

Currently, the app consists of several services which are:

1. `currency-rate` service - the main service of the app, which contains API endpoints and manages fetching rates from third-party services.
2. `email-service` service - manages sending notifications to subscribed users about updates in rates.

Each service has its own tables in the database (currently, single database). Data consistency is guaranteed by a distributed transaction mechanism (orchestration saga).

The whole app has fetches its own metrics and outputs them in a Prometheus-compatible format to [VictoriaMetrics](https://github.com/VictoriaMetrics/VictoriaMetrics).

## Installation Steps

### Local Development

1. Install Golang by following the official installation guide: [Golang Installation Guide](https://golang.org/doc/install)

2. Clone the repository

3. Navigate to the project directory:
    ```bash
    cd genesis-kma-school-entry
    ```

4. Copy the `pkg/settings/.env.sample` file and rename it to `.env` (make sure `.env` file is on the same directory level as `docker-compose.yml`):
    ```bash
    cp ./pkg/settings/.env.sample .env
    ```

5. For each of the services, install the project dependencies:
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

### Unsubscribe to email notifications

- Method: `POST`
- URL: `/unsubscribe`
- Form-data parameter: `email`
- Purpose: unsubscribe from daily email notifications of rate

## Testing

Most of the subpackages are covered by unittests.
In order to run tests:

```bash
go test -v ./...
```

## Metrics

Each service outputs various metrics to VictoriaMetrics, which is configured as Docker Compose container. These include business related metrics, latency (e.g. for database and API endpoints) metrics and other data consistency mechanism (like saga) metrics.

Several ideas for alerts using metrics pushed to VictoriaMetrics and logging are:

1. `currency-rate` service

    1. Basically, any warn or error level log appearing in the log file.
    2. A combination of a recent increase in `saga_total` metric and error log in the compensate event handler signals about a possible inconsistency in data and may require to take immediate action.
    3. Increases in `compensate_total` metrics of any `action` label require an alert and possible future investigation.
    4. `rate_fetcher_consecutive_errors_total` metric going above 5 for any of sources, provided as `fetcher` label. This may indicate that there is some problem with a specific `fetcher` source, whether it be third-party service downtime, eradication of the API key, etc.
    5. And usual hardware metrics and latencies, like drastic increase in `database_*_query_duration_seconds` or `go_cpu_usage` metrics.

2. `email-service` service

    1. `email_errors_total` metric and respective logs signify about some problem in sending prepared emails, most probably with connection and SMTP server.
    2. `notifications_failed_total` metric and respective logs may need to take immediate action, as notifications being the main purpose of this service in particular and the app entirely.
    3. And usual hardware metrics and latencies, like drastic increase in `database_*_query_duration_seconds` or `go_cpu_usage` metrics.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Copyright

© 2024-current Andrii Shalaiev. All Rights Reserved.
