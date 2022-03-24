# aws-go-sdk-test
Sample GO program to use AWS S3

This is just a really quick/simple program to showcase the use of the AWS Go SDK.
It can be use to test your setup and validate your IAM Policies for a user trying to access an S3 bucket.

# Usage

- set an env variable `S3_BUCKET=your-bucket-name` and the AWS Credentials as usual.
- start the app with `./aws-go-sdk-test`

Possible AWS setup:

- using user keys

  ```shell
  export AWS_ACCESS_KEY_ID="xxxxxxx"
  export AWS_SECRET_ACCESS_KEY="yyyyyy"
  ```

- using Assume Role
  This method is the one to use when running the application in an EKS Kubernetes cluster where the ServiceAccount is linked to a IAM Role.

  ```shell
  export AWS_ROLE_ARN=arn:aws:iam::<account>:role/my-aws-go-sdk-role
  export AWS_WEB_IDENTITY_TOKEN_FILE=/var/run/secrets/eks.amazonaws.com/serviceaccount/token
  ```
