# Redis session API example
## Build image
docker build api -t rortegasps/redis-session:1.0.0
docker push rortegasps/redis-session:1.0.0

## Deploy to Kubernetes
### Redis

```bash
kubectl apply -f kubernetes/redis-deployment.yml
kubectl apply -f kubernetes/redis-svc.yml
```

### API

```bash
kubectl apply -f kubernetes/session-api.yml
```

## Test

Get API IP
```bash
export SESSION_API=$(kubectl get svc session-api -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
```

Get profile without sessionID
```bash
curl -X GET \
  http://"$SESSION_API"/profile
```

Login to get a sesionID
```bash
export SESSION_ID=$( \
  curl -X POST \
  http://"$SESSION_API"/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"Hugo","password":"Hugo123"}' \
  | jq -r '.sessionID')
```

Get profile with sessionID
```bash
curl -X GET \
  http://"$SESSION_API"/profile \
  -H 'SessionID: '"$SESSION_ID"''
```
  
Refresh session
```bash
curl -X POST \
  http://"$SESSION_API"/refresh \
  -H 'SessionID: '"$SESSION_ID"''
```

Wait 2.5 minutes (Redis delete session after 2 minutes)
```bash
curl -X GET \
  http://"$SESSION_API"/profile \
  -H 'SessionID: '"$SESSION_ID"''
```

## Scale up
```bash
kubectl scale --replicas=10 deployment session-api
```

Validate origin of every request

```bash
curl -X POST \
  http://"$SESSION_API"/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"Hugo","password":"Hugo123"}'
```

```bash
curl -X POST \
  http://"$SESSION_API"/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"Paco","password":"Paco123"}'
```

```bash
curl -X POST \
  http://"$SESSION_API"/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"Luis","password":"Luis123"}'
```