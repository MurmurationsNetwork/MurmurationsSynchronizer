# MurmurationsSynchronizer
## Trigger the serverless function in local
1. Modify the variables in `hit.go`. `hitCount` means how many times we want to hit to the serverless function. `hitUrl` is the url of the serverless function. `apiKey` is the Bearer Token we set.
   ```
   hitCount := 300
   hitUrl := "http://localhost:3000/api"
   apiKey := ""
   ```
2. Execute the following command.
   ```
   go run hit.go
   ```
## Deployment to k8s
### Production/Staging
1. Add secret, replace "YOUR_KEY" with the correct key
```
kubectl create secret generic synchronizer-job-secret --from-literal="API_SECRET_KEY=YOUR_KEY"
```
