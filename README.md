# Terraform Provider Confluent Schema Registry

Run the following command to build the provider

```shell
make build
```


## Test sample configuration

First, build and install the provider.

```shell
make install
```

Spin up docker compose.

```shell
cd docker_compose
docker compose up
```

Go to `localhost:9021` in your web browser and create topics.

Open another shell and get inside `examples/manage_schema` folder.

```shell
cd examples/manage_schema
```

Finally, run the following commands to initialize the workspace and check the sample output.

```shell
terraform init
terraform plan
```
