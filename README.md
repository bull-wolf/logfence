# logfence

Structured log filtering and redaction proxy for containerized applications.

## Installation

```bash
go install github.com/yourusername/logfence@latest
```

Or build from source:

```bash
git clone https://github.com/yourusername/logfence.git && cd logfence && go build ./...
```

## Usage

Run logfence as a sidecar or standalone proxy that sits between your application and its log output.

```bash
logfence --config logfence.yaml --input stdin --output stdout
```

Example `logfence.yaml`:

```yaml
rules:
  - field: password
    action: redact
  - field: email
    action: mask
  - field: level
    value: debug
    action: drop
```

Pipe your application logs through logfence:

```bash
./myapp 2>&1 | logfence --config logfence.yaml
```

logfence parses structured JSON logs, applies your filtering and redaction rules, and forwards the sanitized output to stdout or a configured destination.

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--config` | `logfence.yaml` | Path to config file |
| `--input` | `stdin` | Log input source |
| `--output` | `stdout` | Log output destination |
| `--format` | `json` | Log format (`json`, `logfmt`) |

## License

MIT © 2024 logfence contributors