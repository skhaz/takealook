# Deploy

```shell
DOCKER_HOST=ssh://pi@192.168.0.19 docker compose --file docker-compose.yaml --file docker-compose.cloudflare.yaml --file docker-compose.logging.yaml up --build --detach
```
