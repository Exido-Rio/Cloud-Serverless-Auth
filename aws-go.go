package main


import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigateway"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssecretsmanager"

	// wrong package installed  github.com/aws/aws-sdk-go-v2/service/apigateway // it's for service not for deployment 
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"

	// "github.com/aws/aws-cdk-go/awscdk/v2/awssqs"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type AwsGoStackProps struct {
	awscdk.StackProps
}

func GenerateRandomSecret(length int) string {
	// Create a byte slice to hold the random bytes
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return ""
	}

	// Encode the random bytes to a base64 string
	// You can also use hex encoding or any other encoding as needed
	secret := base64.StdEncoding.EncodeToString(bytes)
	return secret
}

// off-course you can use must better crypto or other random key-gen with any other lib 

func NewAwsGoStack(scope constructs.Construct, id string, props *AwsGoStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	sec := GenerateRandomSecret(16) // when update it may take some time to make the secret in readable state 

	// create a secret in secretsmanager to use for jwt signing and verification
	secret := awssecretsmanager.NewSecret(stack, jsii.String("MyJWTSecret"), &awssecretsmanager.SecretProps{
		SecretName:        jsii.String("JWTSecret"),
		SecretStringValue: awscdk.SecretValue_UnsafePlainText(&sec),
	})

	// didn't use key enveloping as it would need some extra kms assume etc

	// create a dynamodb table

	table := awsdynamodb.NewTable(stack, jsii.String("myUserTable"), &awsdynamodb.TableProps{
		PartitionKey: &awsdynamodb.Attribute{
			Name: jsii.String("username"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		TableName:     jsii.String("userTable"),
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY, //to be able to remove the DB with cdk destroy
	})

	deployTag := fmt.Sprintf("v-%d", time.Now().Unix())

	myFunction := awslambda.NewFunction(stack, jsii.String("myLambdaFunc"), &awslambda.FunctionProps{
		Runtime: awslambda.Runtime_PROVIDED_AL2023(),
		Code:    awslambda.AssetCode_FromAsset(jsii.String("lambda/function.zip"), nil),
		Handler: jsii.String("main"),
		Environment: &map[string]*string{
			"DEPLOY_TAG": jsii.String(deployTag), // Triggers full environment refresh otherwise used old env which then silently causes runtime error with no logs or error in cloudwatch or anywhere
		},
	})


	// granting the read secret to function

	secret.GrantRead(myFunction, nil)

	table.GrantReadWriteData(myFunction)

	api := awsapigateway.NewRestApi(stack, jsii.String("myApiGateway"), &awsapigateway.RestApiProps{
		DefaultCorsPreflightOptions: &awsapigateway.CorsOptions{
			AllowHeaders: jsii.Strings("Content-Type", "Authorization"),
			AllowMethods: jsii.Strings("GET", "POST", "DELETE", "PUT", "OPTIONS"),
			AllowOrigins: jsii.Strings("*"),
		},
		DeployOptions: &awsapigateway.StageOptions{
			LoggingLevel: awsapigateway.MethodLoggingLevel_INFO,
		},

		CloudWatchRole: jsii.Bool(true),
	})

	integration := awsapigateway.NewLambdaIntegration(myFunction, nil)

	// routes

	registerResource := api.Root().AddResource(jsii.String("register"), nil)
	registerResource.AddMethod(jsii.String("POST"), integration, nil)

	loginResource := api.Root().AddResource(jsii.String("login"), nil)
	loginResource.AddMethod(jsii.String("POST"), integration, nil)

	protectedResource := api.Root().AddResource(jsii.String("protected"), nil)
	protectedResource.AddMethod(jsii.String("GET"), integration, nil)

	return stack

}

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	NewAwsGoStack(app, "AwsGoStack", &AwsGoStackProps{
		awscdk.StackProps{
			Env: env(),
		},
	})

	app.Synth(nil)
}

// env determines the AWS environment (account+region) in which our stack is to
// be deployed. For more information see: https://docs.aws.amazon.com/cdk/latest/guide/environments.html
func env() *awscdk.Environment {
	// If unspecified, this stack will be "environment-agnostic".
	// Account/Region-dependent features and context lookups will not work, but a
	// single synthesized template can be deployed anywhere.
	//---------------------------------------------------------------------------
	return nil

	// Uncomment if you know exactly what account and region you want to deploy
	// the stack to. This is the recommendation for production stacks.
	//---------------------------------------------------------------------------
	// return &awscdk.Environment{
	//  Account: jsii.String("123456789012"),
	//  Region:  jsii.String("us-east-1"),
	// }

	// Uncomment to specialize this stack for the AWS Account and Region that are
	// implied by the current CLI configuration. This is recommended for dev
	// stacks.
	//---------------------------------------------------------------------------
	// return &awscdk.Environment{
	//  Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
	//  Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
	// }
}
