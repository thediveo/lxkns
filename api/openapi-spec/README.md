# lxkns OpenAPI 3.0 Specification

This directory contains the REST API specification of the lxkns service, in
[OpenAPI 3.0 format](https://github.com/OAI/OpenAPI-Specification/).

## Brower-Based Editing

### Swagger

For those of us who are not adverse to editing OpenAPI specifications using a
browser-based UI instead of the plain YAML, there are several options. These are
just two of them, the first one being Swagger's own editor:

```bash
docker run -d --name swagger-editor -p 8080:8080 swaggerapi/swagger-editor
```

### Apicurito

[Apicuri(t)o](https://www.apicur.io/apicurito/) is another editor, featuring
both a visualization as the YAML/JSON definition views. This starts Apicurito
which is the editor component only without any (database) backend.

```bash
docker run -d --name apicurito -p 8080:8080 apicurio/apicurito-ui
```
