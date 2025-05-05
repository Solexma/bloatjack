# BloatJack

A cyber-surgeon that slims your containers by measuring, explaining, and automatically right-sizing resources.

## Features

- 🔍 Real-time container resource profiling
- ⚡ Automatic container optimization
- 🚀 Remote service offloading
- 📊 Resource usage dashboard
- 🔧 Open-source CLI tool

## Installation

### Binary Installation

Download the latest release for your platform from the [Releases](https://github.com/Solexma/bloatjack/releases) page.

```bash
# For macOS (using Homebrew)
brew install Solexma/bloatjack/bloatjack

# For Linux
curl -L https://github.com/Solexma/bloatjack/releases/latest/download/bloatjack-linux-amd64 -o /usr/local/bin/bloatjack
chmod +x /usr/local/bin/bloatjack

# For Windows (using Scoop)
scoop install bloatjack
```

### From Source

```bash
# Clone the repository
git clone https://github.com/Solexma/bloatjack.git
cd bloatjack

# Install dependencies
make deps

# Build
make build

# The binary will be available in bin/bloatjack
```

## Quick Start

```bash
# Scan your containers
bloatjack scan

# Apply optimizations
bloatjack tune

# View dashboard
bloatjack ui
```

## Development

### Prerequisites

- Go 1.24 or later
- Make
- Git Flow

### Setup

```bash
# Clone the repository
git clone https://github.com/Solexma/bloatjack.git
cd bloatjack

# Initialize Git Flow
git flow init

# Install dependencies
make deps

# Run tests
make test

# Build
make build
```

### Building for Distribution

```bash
# Build binaries for all platforms
make dist

# The binaries will be available in dist/
```

### Git Flow Workflow

This project follows Git Flow branching strategy:

- `main` - Production-ready code
- `develop` - Integration branch for features
- `feature/*` - New features
- `release/*` - Release preparation
- `hotfix/*` - Production fixes
- `support/*` - Version support

Common Git Flow commands:

```bash
# Start a new feature
git flow feature start feature-name

# Finish a feature
git flow feature finish feature-name

# Start a release
git flow release start v1.0.0

# Finish a release
git flow release finish v1.0.0

# Start a hotfix
git flow hotfix start hotfix-name

# Finish a hotfix
git flow hotfix finish hotfix-name
```

## Project Structure

```plaintext
bloatjack/
├── docs/                # Documentation
├── internal/            # Private application code
|   ├── cli/             # Cli commands
│   ├── root.go          # definisce rootCmd
│   ├── scan.go          # `bloatjack scan`
│   ├── tune.go          # `bloatjack tune`
│   └── rules.go         # `bloatjack rules`
│   ├── rules/           # loader + rulebook validator
│   │   ├── embed.go     # go:embed *.yml
│   │   ├── parser.go    # Parse([]byte) → []Rule
│   │   └── engine.go    # Apply(serviceStats) → Patch
│   ├── compose/         # parsing/AST of docker‑compose
│   ├── profiler/        # wrapper docker stats
│   └── patch/           # diff YAML / git commit
├── pkg/                 # Public library code
│   └── bloatjack.go
├── main.go              # version, init rootCmd, Execute()
└── go.mod
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
