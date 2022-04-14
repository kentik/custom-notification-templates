# How to Develop and Test New Template

**Templates for custom webhooks are defined with [Go Template Syntax](https://pkg.go.dev/text/template). In order to test them, you can use the provided testing script. It runs both natively on any Go-lang supported platforms or use docker.**

## Testing script

The testing scripts basically performs rendering of all templates located within `templates` directory using a representative set of example payloads injected into them which mimics Kentik Notifications system. Adding a new file or editing existing one with `.tmpl` extension within templates directory and running the test script should be sufficient to verify correctness of the template.

Using double `.json.tmpl` enables additional JSON validation of the output content.

### The output directory

Testing script stores rendered notifications within the output directory. It can be helpful to examine these file to verify the contents of notifications will have the expected shape.

## Run test script locally

1. Make sure to have up-to-date Go runtime setup.
2. Run tests using command:

   ```shell
   go test ./pkg
   ```

## Run test script within docker

Running within docker allows you to make the same validation without bothering about Go setup on your host machine:

1. Make sure to have up-to-date Docker environment
2. Run tests using example command:

  ```shell
  docker run -v $PWD:/go/src/app -w /go/src/app golang:1.17 go test ./pkg
  ```

Please note running docker command won't update `output` directory with rendered files as the docker container.
