# file-management-api
Lambda REST API for files managed in the larger 281p1 application 

## Deployment
Deployed via serverless framework dashboard
```shell
// after setting up serverless account, creating app, etc.
$ serverless login
$ make deploy
```

## Middleware
```
JWT Token Authorization
Checks the Access Token's claims agains the resource being requested to ensure that the requestor is authorized to access that information.
```

## APIGateway Endpoints/Lambdas
```
Endpoint: /user-id/file-id
Description: Get, delete, and modify particular files
HTTP Methods: GET, PATCH, DELETE
Authorization: Admin, User
Middleware:
```

```
Endpoint: /user-id
Description: Get files associated with a particular user, or upload a file for that user
HTTP Methods: GET, POST
Authorization: Admin, User
```

```
Endpoint: /
Description: Get all files that are uploaded
HTTP Methods: GET 
Authorization: Admin
```

## Other Lambda Functions
```
Endpoint: None
Usage: Lambda@Edge Function for Cloudfront Distribution for limiting access to distribution to authenticated users.
Implementation:
    Decodes JWT, and checks for correct user authorizations against Cognito Identity Pool (Federated Identities)
```

## Architecture
