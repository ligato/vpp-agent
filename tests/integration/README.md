# Integration Tests

## VPP

The `tests/integration/vpp` directory contains test cases for testing VPP integration.
These tests are using handlers from vppcalls directly. To quickly check VPP behaviour
and avoid regressions when updating VPP.

### Run tests

Quickest and simplest way to run VPP integration tests:

```sh
# Run for default VPP version
make integration-tests

# Run for different VPP version
make integration-tests VPP_VERSION=2005
```
