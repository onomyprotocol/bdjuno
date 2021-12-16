# How the run the bdjuno + hasura with docker compose

* Start postgres

```
docker-compose up -d postgres
```

* Apply scripts from the [schema folder](database/schema)

* Create .bdjuno folder and copy config.toml (based on config-sample.yaml) and genesis.json files. 
  
* Start the bdjuno

```
docker-compose up -d bdjuno
```

* Start the hasura

```
docker-compose up -d hasura
```

* Apply the hasura metadata

```
docker-compose exec hasura hasura metadata apply --admin-secret "myadminsecretkey"
```

The env HASURA_GRAPHQL_ADMIN_SECRET from the dockerfile contains the secret.

If the output is

```
INFO Metadata applied 
```

Then metadata applied. If empty or error, something went wrong.