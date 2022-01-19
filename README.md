# Lacework AWS Security Hub Integration CloudFormation Deployment

## Overview
With Lacework AWS Security Hub integration, you can send Lacework Security Alerts to AWS Security Hub. This repo contains the CloudFormation template to automatically enable this feature.

## Prerequisites
You need the following prerequisites to implement the Lacework AWS Security Hub Integration.

- AWS Security Hub enabled on your AWS account.
- A Lacework Cloud Security Platform SaaS account.

## Installing the Lacework AWS Security Hub Integration

### 1. Generate a Lacework API Access Key

1. In your console, go to **Settings > API Keys**.
2. Click on the **Create New** button in the upper right to create a new API key.
3. Provide a **name** and **description** and click Save.
4. Click the download button to download the API keys file.
5. Copy the **keyId** and **secret** from this file.

### 2. Deploy the Lacework AWS Security Hub Integration with CloudFormation
1. Login in to AWS master account with Administrator permissions.
Click on the following Launch Stack button to go to your CloudFormation console and launch the AWS Control Integration template.
   
   [![Launch Stack](https://user-images.githubusercontent.com/6440106/150169828-1692c426-ce7a-4ee9-ae6e-0a0b2d9a99e8.png)](https://console.aws.amazon.com/cloudformation/home?#/stacks/create/review?templateURL=https://lacework-alliances.s3.us-west-2.amazonaws.com/lacework-aws-security-hub/templates/aws-security-hub-integration.yml)

   For most deployments, you only need the Basic Configuration parameters. Use the Advanced Configuration for customization.
   ![CloudFormation Stack Form](https://user-images.githubusercontent.com/6440106/149715371-62f7f918-ac94-4c6e-8c9d-a8049eda6f9b.png)
3. Specify the following Basic Configuration parameters:
    * Enter a **Stack name** for the stack.
    * Enter **Your Lacework URL**.
    * Enter your **Lacework Sub-Account Name** if you are using Lacework Organizations.
    * Enter your **Lacework Access Key ID** and **Secret Key** that you copied from your previous API Keys file.
    * Choose whether you want to **Create Lacework Alert Channel**. This will create the Lacework alert channel and rule.
    * Enter the **Alert Channel Name**.
4. Click **Next** through to your stack **Review**.
5. Accept the AWS CloudFormation terms and click **Create stack**.

### 3. CloudFormation Progress

1. Monitor the progress of the CloudFormation deployment. It takes several minutes for the stack to create the resources that enable the Lacework AWS Control Tower Integration.
2. When successfully completed, the stack shows CREATE_COMPLETE.

### 4. Validate the Lacework AWS Security Hub Integration

1. Login to your Lacework Cloud Security Platform console.
2. Go to **Settings > Alert Channels**.
3. You should see the new alert channel in the list.
4. Go to **Settings > Alert Rules**.
5. You should see the new alert rule in the list.

## Remove the Lacework AWS Security Hub Integration

To remove the Lacework AWS Security Hub Integration, simply delete the main stack. All CloudFormation stacksets, stack instances, and Lambda functions will be deleted. **Note:** Lacework will no longer send alerts to AWS Security Hub.