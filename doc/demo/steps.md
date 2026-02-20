## Setup kind
```
> kind delete cluster --name k8s-local
> cd cnpg-playground
> ./scripts/setup.sh local
```

## Setup cnpg
```
> cd cnpg-playground
> REGIONS="local" TRUNK="true" ./demo/setup.sh
> kind get kubeconfig --name k8s-local >> /tmp/k8s-local.yaml
> export KUBECONFIG=/tmp/k8s-local.yaml
```

## Remaining steps from local directory
```
> cd cnpg-i-mcp-demo
```

## Apply demo sql data to local cluster
```
> kubectl exec -it pg-local-1 -- psql -U postgres -d app < doc/demo/data.sql
```

## Explore the local data 
```
> kubectl exec -it pg-local-1 -- psql -d app -U postgres
    \d
    select * from employees;
    select * from departments;
```

## Build and load the image
```
> docker build -t cnpg-i-mcp-demo:latest .
> kind load docker-image --name k8s-local cnpg-i-mcp-demo:latest
```

## Deply the plugin
```
> kubectl apply -k kubernetes/ 
> k apply -f doc/examples/cluster-example.yaml
```

## Patch the cluster
```
plugins:
  - name: postgres-mcp-cnpg-plugin 
```

## Connect to the mcp port using claude
```
> claude mcp add postgres-local --transport sse http://localhost:8080/sse # maybe add /sse or /mcp or /mcp/sse to the end 
```

## Ask claude about the data in the database
```
How many customers do I have in postgres-local?
```