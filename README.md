# Log Collection Service

## Local development

- Install Docker
- Run Docker compose using Makefile to start a setup that has 1 primary server and 2 secondary servers
  ```
    make run
  ```
- Run Tests using Makefile:
  ```
    make test
  ```

## API

- Endpoint: `/api/v1/logs` 
- Parameters:
  - n: number of log entries to retrieve
  - file: name of log file
  - filter: basic keyword match filter

Example curl command:

```
curl 'localhost:3000/api/v1/logs?n=10&file=wifi.log&filter=notification'
```
