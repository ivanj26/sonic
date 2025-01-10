# Sonic

Sonic is a high-performance tool designed for seamless slot migration in Redis clusters. It enables parallel operations, ensuring minimal downtime and efficient resharding of keys between nodes. Built with simplicity and reliability in mind, Sonic helps manage Redis cluster operations with ease.

---

## Features

- **Parallel Slot Migration**: Perform concurrent migrations to minimize operation time.
- **Redis Cluster Compatibility**: Fully supports Redis cluster architecture.
- **Error Handling**: Handles common Redis `MIGRATE` issues such as authentication, `NOKEY`, and `BUSYKEY`.
- **Customizable Configuration**: Easily adjust migration parameters such as timeouts and authentication.
- **Logging**: Provides detailed logs for tracking migration progress and troubleshooting.
- **Retry Mechanism**: Automatically retries migration for busy keys or transient errors.

---

## Sonic Installation
1. Build the application using the following command:

    ```bash
    make build
    ```

2. Run the application by using the following command and options:

    ```bash
    ./sonic -s <source_ip_without_port> -d <dest_ip_without_port> -a <password>> -n <slots>           
    ```

## Usage

Sonic provides an intuitive interface to manage slot migrations.

### Command-Line Arguments

| Argument         | Description                                   | Default          | Example
|------------------|-----------------------------------------------|------------------|------------------
| `-s`       | Address of the source Redis node without port, default port is 6379. Should be a master node.             | N/A              | 10.21.123.12
| `-d`  | Address of the destination Redis node without port, default ort is 6379. Should be a master node.       | N/A              | 10.21.123.13
| `-n`        | Comma-separated list of slots to migrate.     | N/A              | 1365,1380 or 1365 (single slot)
| `-a`        | Redis password.     | N/A              |
| `-p`     | Number of parallel migrations to execute. By default: 5 at max concurrent     | `5`              | 5