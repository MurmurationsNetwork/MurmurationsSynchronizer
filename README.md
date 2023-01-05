# MurmurationsSynchronizer
## Trigger the serverless function in local
1. Modify the variables in `hit.go`. `hitCount` means how many times we want to hit to the serverless function. `hitUrl` is the url of the serverless function.
   ```
   hitCount := 300
   hitUrl := "http://localhost:3000/api"
   ```
2. Execute the following command.
   ```
   go run hit.go
   ```