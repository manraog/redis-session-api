# Redis session API example
## Build image
docker build api -t rortegasps/redis-session:1.0.0

## Deploy to Kubernetes
### Redis
kubectl apply -f kubernetes/redis-deployment.yml
kubectl apply -f kubernetes/redis-svc.yml

### API
kubectl apply -f kubernetes/session-api.yml

## Test
Get API IP
export SESSION_API=$(kubectl get svc redis-session-api -o jsonpath='{.status.loadBalancer.ingress[0].ip}')

curl -X GET \
  http://"$SESSION_API"/profile

curl -X POST \
  http://"$SESSION_API"/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"Hugo","password":"Hugo123"}'

curl -X GET \
  http://"$SESSION_API"/profile \
  -H 'SessionID: d1762368-4ec2-453a-a77d-b11125ea4f14'
  
curl -X POST \
  http://"$SESSION_API"/refresh \
  -H 'SessionID: d1762368-4ec2-453a-a77d-b11125ea4f14'

## Scale up
kubectl scale --replicas=10 deployment session-api
