# Elasticsearch + Postgresql + Golang

## Description

Searching products with Suggestion using Elasticsearch + Postgresql + Golang

## How To

1. Run **make dev**
2. Run in terminal to create superuser password in Elasticsearch :

```
docker exec -it es bin/elasticsearch-setup-passwords interactive
```

3. Run this command in terminal to create role and user. Use your own role, username and password

```
curl -u elastic:es12345 -X POST "http://localhost:9200/_security/role/ecs_own_role" \
  -H "Content-Type: application/json" \
  -d '{
    "indices": [
      {
        "names": ["products"],
        "privileges": ["read", "write", "create", "delete", "create_index", "view_index_metadata"]
      }
    ]
  }'
```

```
curl -u elastic:es12345 -X POST "http://localhost:9200/_security/user/ecsapp" \
  -H "Content-Type: application/json" \
  -d '{
    "password" : "ecs123",
    "roles" : ["ecs_own_role"],
    "full_name" : "ECS App User"
  }'
```

4. Open Postman and use http://localhost:8080/health
5. All routes in main.go

## Contact

For further inquiries, please contact the development team.
