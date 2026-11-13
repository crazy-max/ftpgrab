# Integration tests

Run from the repository root with Docker Compose:

```console
docker compose -f test/compose.yaml up --build --abort-on-container-exit --exit-code-from tests
```

Clean up afterward:

```console
docker compose -f test/compose.yaml down --volumes --remove-orphans
```
