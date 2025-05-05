# Contributing to BloatJack

Thank you for your interest in contributing to BloatJack! This document provides guidelines and instructions for contributing.

## Code of Conduct

Please read and follow our [Code of Conduct](CODE_OF_CONDUCT.md).

## How to Contribute

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Development Setup

1. Install Go 1.24 or later
2. Clone the repository
3. Run `make deps` to install dependencies
4. Run `make test` to verify your setup

## Pull Request Process

1. Update the README.md with details of changes if needed
2. Update the CHANGELOG.md with your changes
3. The PR will be merged once you have the sign-off of at least one maintainer

## Adding New Rules

When adding new rules:

1. Place them in the appropriate YAML file in `internal/rules/`
2. Follow the existing rule format
3. Add tests for your rules
4. Update the VERSION file

## Reporting Bugs

Please use the GitHub issue tracker to report bugs. Include:

- A clear description of the problem
- Steps to reproduce
- Expected behavior
- Actual behavior
- Environment details

## License

By contributing, you agree that your contributions will be licensed under the project's MIT License.
