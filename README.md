## Setup local development

### Install tools

- [Docker desktop](https://www.docker.com/products/docker-desktop)
- [TablePlus](https://tableplus.com/)
- [Golang](https://golang.org/)
- [Homebrew](https://brew.sh/)
- [Migrate](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate)

    ```bash
    brew install golang-migrate
    ```

- [DB Docs](https://dbdocs.io/docs)

    ```bash
    npm install -g dbdocs
    dbdocs login
    ```

- [DBML CLI](https://www.dbml.org/cli/#installation)

    ```bash
    npm install -g @dbml/cli
    dbml2sql --version
    ```

- [Sqlc](https://github.com/kyleconroy/sqlc#installation)

    ```bash
    brew install sqlc
    ```

- [Gomock](https://github.com/golang/mock)

    ``` bash
    go install github.com/golang/mock/mockgen@v1.6.0
    ```

- [Velero](https://velero.io/docs/v1.13/basic-install/)
    ``` bash
    brew install velero
    ```

### Setup infrastructure

- Create the bank-network

    ``` bash
    make network
    ```

- Start postgres container:

    ```bash
    make postgres
    ```

- Create simple_bank database:

    ```bash
    make createdb
    ```

- Run db migration up all versions:

    ```bash
    make migrateup
    ```

- Run db migration up 1 version:

    ```bash
    make migrateup1
    ```

- Run db migration down all versions:

    ```bash
    make migratedown
    ```

- Run db migration down 1 version:

    ```bash
    make migratedown1
    ```

### Documentation

- Generate DB documentation:

    ```bash
    make db_docs
    ```

- Access the DB documentation at [this address](https://dbdocs.io/techschool.guru/simple_bank). Password: `secret`

### How to generate code

- Generate schema SQL file with DBML:

    ```bash
    make db_schema
    ```

- Generate SQL CRUD with sqlc:

    ```bash
    make sqlc
    ```

- Generate DB mock with gomock:

    ```bash
    make mock
    ```

- Create a new db migration:

    ```bash
    make new_migration name=<migration_name>
    ```

### How to run

- Run server:

    ```bash
    make server
    ```

- Run test:

    ```bash
    make test
    ```

## Deploy to kubernetes cluster

- [Install nginx ingress controller](https://kubernetes.github.io/ingress-nginx/deploy/#aws):

    ```bash
    kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v0.48.1/deploy/static/provider/aws/deploy.yaml
    ```

- [Install cert-manager](https://cert-manager.io/docs/installation/kubernetes/):

    ```bash
    kubectl apply -f https://github.com/jetstack/cert-manager/releases/download/v1.4.0/cert-manager.yaml
    ```


## Working with multiple K8 clusters
- To get AWS to create a kubeconfig  use the following command:

    ```bash
    aws eks update-kubeconfig --name simple-bank-1 --region us-west-1
    ```
- To view configs
    ```bash
    kubectl config view
    ```
- To switch to a cluster
    ```bash
    kubectl config use-context <cluster name>
    ```

## Backing up Kubernates cluster info to S3 via velero 
You can backup all the cluster information to S3 via velero and delete the cluster to keep AWS EKS costs down. 
- Install velero
- Config the AWS S3 plugin as shown here https://github.com/vmware-tanzu/velero-plugin-for-aws#setup
- Once the S3 buckets have been created and permissions granted (step 2) save the configs to a backup using the following command
    ```bash 
    velero backup create simple-bank-k8-backup --include-namespaces default,cert-manager --kubecontext <context name>
    ```
- Delete the node group and cluster in EKS
- When you want to restart, create a new EKS cluster and node group, and then install velero controllers $BUCKET is the S3 bucket and $REGION is the AWS region.
    ```bash
    velero install \
    --provider aws \
    --plugins velero/velero-plugin-for-aws:v1.9.0 \
    --bucket $BUCKET \
    --backup-location-config region=$REGION \
    --snapshot-location-config region=$REGION \
    --secret-file ./credentials-velero
    ```
6. Restore the backup using the following command
    ```bash
    velero restore create --from-backup simple-bank-k8-backup --kubecontext <context name>
    ```
7. Add the elb address from the ingress object to the DNS A record in Route53.
