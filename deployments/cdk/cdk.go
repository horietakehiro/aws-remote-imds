package main

import (
	"fmt"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscodebuild"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscodepipeline"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscodepipelineactions"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslogs"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"

	// "github.com/aws/aws-cdk-go/awscdk/v2/awssqs"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type CdkStackProps struct {
	awscdk.StackProps
}

const (
	APP_NAME = "aws-remote-imds"
)

func NewCdkStack(scope constructs.Construct, id string, props *CdkStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// The code that defines your stack goes here

	// example resource
	// queue := awssqs.NewQueue(stack, jsii.String("CdkQueue"), &awssqs.QueueProps{
	// 	VisibilityTimeout: awscdk.Duration_Seconds(jsii.Number(300)),
	// })

	return stack
}

func CiCdStach(scope constructs.Construct, id string, props *CdkStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	pipelinRole := awsiam.NewRole(stack, jsii.String("PipelineRole"), &awsiam.RoleProps{
		AssumedBy: awsiam.NewServicePrincipal(jsii.String("codepipeline.amazonaws.com"), nil),
		RoleName:  jsii.String(fmt.Sprintf("%s-cicd-pipeline-role", APP_NAME)),
	})
	pipelinRole.AddManagedPolicy(awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("AdministratorAccess")))
	buildRole := awsiam.NewRole(stack, jsii.String("BuildRole"), &awsiam.RoleProps{
		AssumedBy: awsiam.NewServicePrincipal(jsii.String("codebuild.amazonaws.com"), nil),
		RoleName:  jsii.String(fmt.Sprintf("%s-cicd-build-role", APP_NAME)),
	})
	buildRole.AddManagedPolicy(awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("AdministratorAccess")))

	pipeline := awscodepipeline.NewPipeline(stack, jsii.String("Pipeline"), &awscodepipeline.PipelineProps{
		ArtifactBucket: awss3.Bucket_FromBucketName(
			stack, jsii.String("PrivateartifacatBucket"), jsii.String("private-artifact-bucket-382098889955-ap-northeast-1"),
		),
		PipelineName:             jsii.String(fmt.Sprintf("%s-cicd-pipeline", APP_NAME)),
		RestartExecutionOnUpdate: jsii.Bool(false),
		Role:                     pipelinRole,
	})
	sourceArtifact := awscodepipeline.NewArtifact(jsii.String("SourceArtifact"))
	buildArtifact := awscodepipeline.NewArtifact(jsii.String("BuildArtifact"))
	githubSourceAction := awscodepipelineactions.NewCodeStarConnectionsSourceAction(
		&awscodepipelineactions.CodeStarConnectionsSourceActionProps{
			ActionName:         jsii.String("Source"),
			RunOrder:           jsii.Number(1),
			VariablesNamespace: jsii.String("SourceVariables"),
			Role:               pipelinRole,
			ConnectionArn:      jsii.String("arn:aws:codestar-connections:ap-northeast-1:382098889955:connection/26404591-2de4-4d56-acd0-93232fcdfb27"),
			Repo:               jsii.String(APP_NAME),
			Branch:             jsii.String("dev"),
			TriggerOnPush:      jsii.Bool(true),
			Output:             sourceArtifact,
			Owner:              jsii.String("horietakehiro"),
		},
	)

	buildProject := awscodebuild.NewPipelineProject(stack, jsii.String("BuildProject"), &awscodebuild.PipelineProjectProps{
		BuildSpec: awscodebuild.BuildSpec_FromSourceFilename(jsii.String("./buildspec.yaml")),
		Environment: &awscodebuild.BuildEnvironment{
			BuildImage:           awscodebuild.LinuxBuildImage_AMAZON_LINUX_2_4(),
			ComputeType:          awscodebuild.ComputeType_SMALL,
			Privileged:           jsii.Bool(true),
			EnvironmentVariables: &map[string]*awscodebuild.BuildEnvironmentVariable{},
		},
		GrantReportGroupPermissions: jsii.Bool(true),
		ProjectName:                 jsii.String(fmt.Sprintf("%s-cicd-build-project", APP_NAME)),
		Logging: &awscodebuild.LoggingOptions{
			CloudWatch: &awscodebuild.CloudWatchLoggingOptions{
				Enabled: jsii.Bool(true),
				LogGroup: awslogs.NewLogGroup(stack, jsii.String("BuildLogGroup"), &awslogs.LogGroupProps{
					Retention:    awslogs.RetentionDays_FIVE_DAYS,
					LogGroupName: jsii.String(fmt.Sprintf("%s-cicd-build-project", APP_NAME)),
				}),
			},
		},
		Role: buildRole,
	})
	buildAction := awscodepipelineactions.NewCodeBuildAction(
		&awscodepipelineactions.CodeBuildActionProps{
			ActionName:                          jsii.String("Build"),
			RunOrder:                            jsii.Number(1),
			VariablesNamespace:                  jsii.String("BuildVariables"),
			Role:                                buildRole,
			Input:                               sourceArtifact,
			CheckSecretsInPlainTextEnvVariables: jsii.Bool(false),
			// EnvironmentVariables:                map[string]awscodebuild.BuildEnvironmentVariable{},
			Project: buildProject,
			Outputs: &[]awscodepipeline.Artifact{
				buildArtifact,
			},
		},
	)

	pipeline.AddStage(&awscodepipeline.StageOptions{
		StageName:           jsii.String("Source"),
		TransitionToEnabled: jsii.Bool(true),
		Actions: &[]awscodepipeline.IAction{
			githubSourceAction,
		},
	})
	pipeline.AddStage(&awscodepipeline.StageOptions{
		StageName:           jsii.String("Build"),
		TransitionToEnabled: jsii.Bool(true),
		Actions: &[]awscodepipeline.IAction{
			buildAction,
		},
	})

	return stack
}

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	CiCdStach(app, "CiCdStach", &CdkStackProps{
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
