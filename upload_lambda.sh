#!/usr/bin/env bash

ACCOUNT_ID=''
LAMBDA_NAME='start-stop-ec2'
ROLE_NAME='start-stop-ec2'
ROLE_ARN="arn:aws:iam::${ACCOUNT_ID}:role/${ROLE_NAME}"

echo "building lambda $LAMBDA_NAME"
echo "building go program and creating .zip file for upload to AWS Lambda"

env GOOS=linux GOARCH=amd64 go build -o start-ec2
zip -j start-ec2.zip start-ec2

aws iam get-role --role-name ${ROLE_NAME}
if [[ $? -eq 0 ]]; then
    echo "$ROLE_NAME exists, proceeding to skip creation"
else
    echo "$ROLE_NAME does not exist, creating IAM Role...."
    aws iam create-role --role-namee start-stop-ec2 \
    --asume-role-policy-document file://lambda-policy.json

    aws iam attach-role-policy --role-name start-stop-ec2 \
    --policy-arn arn:aws:iam:aws:policy/service-role/AWSLsmbdaBasicExecutionRole
fi

echo "deleting current lambda function if found; lambda: $LAMBDA_NAME"
aws lambda get-function --function-name ${LAMBDA_NAME} | grep "FunctionArn"
if [[ $? -eq 0 ]]; then
    echo "lambda ${LAMBDA_NAME} exists, deleting in order to re-create and upload new version"
    aws lambda delete-function --function-name ${LAMBDA_NAME}
fi

echo "uploading new lambda...."
aws lambda create-function \
--function-name ${LAMBDA_NAME} \
--runtime go1.x \
--timeout 30 \
--role ${ROLE_ARN} \
--zip-file fileb://start-ec2.zip \
--handler start-ec2 \
--vpc-config SubnetIds=subnet-{{enter list}},SecurityGroupIds={{enter list}} \
--memory-size 128 \
--publish

# cleanup
rm start-ec2 && rm start-ec2.zip
