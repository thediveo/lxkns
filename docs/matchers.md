# Gomega Matchers

Developers using the `lxkns` Go API (as opposed to the service REST API) might
be interested in dedicated [Gomega](https://onsi.github.io/gomega/) support.

The `matcher/` directory contains Gomega matchers for matching container
names/IDs (and optional their types and flavors), container group memberships,
and more.

```go
var containers []*model.Container

Expect(containers).To(ContainElement(HaveContainerName("foobar")))
```
