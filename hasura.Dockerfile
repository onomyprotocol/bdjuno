FROM hasura/graphql-engine:v2.0.4
RUN apt-get update && apt-get install -y curl
RUN curl -L https://github.com/hasura/graphql-engine/raw/stable/cli/get.sh | bash
COPY hasura hasura
WORKDIR hasura

CMD ["graphql-engine", "serve"]