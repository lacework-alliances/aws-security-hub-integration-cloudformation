package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-lambda-go/cfn"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/jefferyfry/funclog"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
)

var (
	LogI = funclog.NewInfoLogger("INFO: ")
	LogW = funclog.NewInfoLogger("WARN: ")
	LogE = funclog.NewErrorLogger("ERROR: ")
)

type PostAccessTokenRequest struct {
	KeyId string `json:"keyId"`
	ExpiryTime int `json:"expiryTime"`
}

type PostAccessTokenResponse struct {
	ExpiresAt string `json:"expiresAt"`
	Token string `json:"token"`
}

type PostAlertChannelRequest struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Enabled int `json:"enabled"`
	Data PostAlertChannelRequestData `json:"data"`
}

type PostAlertChannelRequestData struct {
	IssueGrouping string `json:"issueGrouping"`
	EventBusArn string `json:"eventBusArn"`
}

type PostAlertChannelResponse struct {
	Data PostAlertChannelResponseData `json:"data"`
}

type PostAlertChannelResponseData struct {
	CreatedOrUpdatedBy string `json:"createdOrUpdatedBy"`
	CreatedOrUpdatedTime string `json:"createdOrUpdatedTime"`
	Enabled int `json:"enabled"`
	IntgGuid string `json:"intgGuid"`
	IsOrg int `json:"isOrg"`
	Name string `json:"name"`
	Props json.RawMessage `json:"props"`
	State json.RawMessage `json:"state"`
	Type string `json:"type"`
	Data json.RawMessage `json:"data"`
}

type PostAlertRuleRequest struct {
	Filters FiltersArray `json:"filters"`
	IntgGuidList []string `json:"intgGuidList"`
	Type string `json:"type"`
}

type FiltersArray struct {
	Name string `json:"name"`
	Description string `json:"description"`
	Enabled int `json:"enabled"`
	ResourceGroups []string `json:"resourceGroups"`
	EventCategory []string `json:"eventCategory"`
	Severity []int `json:"severity"`
}

type PostHoneycombRequest struct {
	Account string `json:"account"`
	SubAccount string `json:"sub-account"`
	TechPartner string `json:"tech-partner"`
	IntegrationName string `json:"integration-name"`
	Version string `json:"version"`
	Service string `json:"service"`
	InstallMethod string `json:"install-method"`
	Function string `json:"function"`
	Event string `json:"event"`
	EventData string `json:"event-data"`
}

//create the alert channel with
func main() {
	lambda.Start(cfn.LambdaWrap(handler))
}

func handler(ctx context.Context, event cfn.Event) (physicalResourceID string, data map[string]interface{}, err error) {
	if event.RequestType == cfn.RequestCreate {
		return create(ctx,event)
	} else if event.RequestType == cfn.RequestDelete {
		return delete(ctx,event)
	} else {
		LogW.Println("CloudFormation event not supported: ",event.RequestType)
		return "",nil,nil
	}
}

func create(ctx context.Context, event cfn.Event) (physicalResourceID string, data map[string]interface{}, err error) {
	LogI.Printf("CloudFormation event received: %+v \n",event)
	laceworkUrl := os.Getenv("lacework_url")
	subAccountName := os.Getenv("lacework_sub_account_name")
	accessKeyId := os.Getenv("lacework_access_key_id")
	secretKey := os.Getenv("lacework_secret_key")
	eventBusArn := os.Getenv("event_bus_arn")
	alertChannelName := os.Getenv("alert_channel_name")
	sendHoneycombEvent(strings.Split(laceworkUrl,".")[0],"create started",subAccountName,"{}")

	cfnResponse := cfn.NewResponse(&event)

	valid := true

	if laceworkUrl == "" {
		LogE.Println("laceworkUrl was not set.")
		valid = false
	}

	if subAccountName == "" {
		LogW.Println("laceworkSubAccountName was not set.")
	}

	if accessKeyId == "" {
		LogE.Println("laceworkAccessKeyId was not set.")
		valid = false
	}

	if secretKey == "" {
		LogE.Println("laceworkSecretKey was not set.")
		valid = false
	}

	if eventBusArn == "" {
		LogE.Println("eventBusArn was not set.")
		valid = false
	}

	if !valid {
		cfnResponse.Status = cfn.StatusFailed
		cfnResponse.Reason = "Required environment variables were not set."
		cfnResponse.Send()
		return event.PhysicalResourceID,nil,errors.New("unable to run setup due to missing required environment variables")
	}

	LogI.Println("Getting access token.")
	if accessToken, err := sendPostAccessTokenRequest(laceworkUrl,accessKeyId,secretKey); err == nil {
		LogI.Println("Creating Alert Channel.")
		if intgGuid, err := sendPostAlertChannelRequest(alertChannelName,eventBusArn,laceworkUrl,*accessToken,subAccountName); err == nil {
			LogI.Println("Creating Alert Rule.")
			if err := sendPostAlertRuleRequest(alertChannelName,*intgGuid,laceworkUrl,*accessToken,subAccountName); err != nil {
				errMsg := fmt.Sprintf("Failed creating alert rule: %v",err)
				LogE.Println(errMsg)
				cfnResponse.Status = cfn.StatusFailed
				cfnResponse.Reason = errMsg
				cfnResponse.Send()
				return event.PhysicalResourceID,nil,err
			}
		} else {
			errMsg := fmt.Sprintf("Failed creating alert channel: %v",err)
			LogE.Println(errMsg)
			cfnResponse.Status = cfn.StatusFailed
			cfnResponse.Reason = errMsg
			cfnResponse.Send()
			return event.PhysicalResourceID,nil,err
		}
	} else {
		errMsg := fmt.Sprintf("Failed creating access token: %v",err)
		LogE.Println(errMsg)
		cfnResponse.Status = cfn.StatusFailed
		cfnResponse.Reason = errMsg
		cfnResponse.Send()
		return event.PhysicalResourceID,nil,err
	}

	sendHoneycombEvent(strings.Split(laceworkUrl,".")[0],"create completed",subAccountName,"{}")
	cfnResponse.Status = cfn.StatusSuccess
	cfnResponse.Send()
	return event.PhysicalResourceID,nil,nil
}

func delete(ctx context.Context, event cfn.Event) (physicalResourceID string, data map[string]interface{}, err error) {
	LogI.Printf("CloudFormation event received: %+v \n",event)
	laceworkUrl := os.Getenv("lacework_url")
	subAccountName := os.Getenv("lacework_sub_account_name")
	sendHoneycombEvent(strings.Split(laceworkUrl,".")[0],"delete started",subAccountName,"{}")

	cfnResponse := cfn.NewResponse(&event)
	/**
	TODO: Delete alert channel, delete alert rule
	 */
	sendHoneycombEvent(strings.Split(laceworkUrl,".")[0],"delete completed",subAccountName,"{}")
	cfnResponse.Status = cfn.StatusSuccess
	cfnResponse.Send()
	return event.PhysicalResourceID,nil,nil
}

func sendPostAccessTokenRequest(laceworkUrl string, accessKeyId string, secretKey string) (*string, error) {
	requestPayload := PostAccessTokenRequest {
		KeyId: accessKeyId,
		ExpiryTime: 86400,
	}
	if payloadBytes, err := json.Marshal(requestPayload); err == nil {
		request, err := http.NewRequest(http.MethodPost, "https://" + laceworkUrl + "/api/v2/access/tokens", bytes.NewBuffer(payloadBytes))

		if err != nil {
			return nil, err
		}

		request.Header.Add("X-LW-UAKS", secretKey)
		request.Header.Add("content-type", "application/json")

		if resp, err  := http.DefaultClient.Do(request); err == nil {
			defer resp.Body.Close()
			respData := PostAccessTokenResponse{}
			if err := json.NewDecoder(resp.Body).Decode(&respData); err == nil {
				LogI.Printf("PostAccessTokenResponse: %+v",respData)
			} else {
				LogE.Printf("Unable to get response body: %v",err)
				return nil, err
			}
			if resp.StatusCode == http.StatusCreated {
				return &respData.Token, nil
			} else {
				LogE.Printf("Failed to get access token %d",resp.StatusCode)
				return nil,errors.New(string(resp.StatusCode))
			}
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

func sendPostAlertChannelRequest(name string, eventBusArn string, laceworkUrl string, accessToken string, subAccountName string) (*string, error) {
	requestPayload := PostAlertChannelRequest {
		Name: name,
		Type: "CloudwatchEb",
		Enabled: 1,
		Data: PostAlertChannelRequestData {
			IssueGrouping: "Events",
			EventBusArn: eventBusArn,
		},
	}
	if payloadBytes, err := json.Marshal(requestPayload); err == nil {
		if resp, err  := sendApiPostRequest(laceworkUrl,"/api/v2/AlertChannels",accessToken,payloadBytes,subAccountName); err == nil {
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusCreated {
				respData := PostAlertChannelResponse{}
				if err := json.NewDecoder(resp.Body).Decode(&respData); err == nil {
					respDump, err := httputil.DumpResponse(resp, true)
					if err != nil {
						LogW.Println(err)
					}
					LogI.Printf("Received response: %s", string(respDump))
					return &respData.Data.IntgGuid, nil
				} else {
					LogE.Printf("Unable to get response body: %v",err)
					return nil, err
				}
			} else {
				return nil,errors.New(fmt.Sprintf("Response status is %d",resp.StatusCode))
			}

		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

func sendPostAlertRuleRequest(name string,intgGuid string, laceworkUrl string, accessToken string, subAccountName string) error {
	requestPayload := PostAlertRuleRequest {
		Filters: FiltersArray {
			Name: name,
			Description: "Alert rule for Lacework AWS Security Hub",
			Enabled: 1,
			ResourceGroups: []string{},
			EventCategory: []string{},
			Severity: []int{1,2,3,4,5},
		},
		IntgGuidList: []string{intgGuid},
		Type: "Event",
	}
	if payloadBytes, err := json.Marshal(requestPayload); err == nil {
		if resp, err  := sendApiPostRequest(laceworkUrl,"/api/v2/AlertRules",accessToken,payloadBytes,subAccountName); err == nil {
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusCreated {
				respDump, err := httputil.DumpResponse(resp, true)
				if err != nil {
					LogW.Println(err)
				}
				LogI.Printf("Received response: %s", string(respDump))
				return nil
			} else {
				return errors.New(fmt.Sprintf("Response status is %d",resp.StatusCode))
			}
		} else {
			return err
		}
	} else {
		return err
	}
}

func sendApiPostRequest(laceworkUrl string, api string, accessToken string, requestPayload []byte, subAccountName string) (*http.Response, error) {
	request, err := http.NewRequest(http.MethodPost, "https://" + laceworkUrl + api, bytes.NewBuffer(requestPayload))

	if err != nil {
		LogE.Printf("Error creating API post request: %v %v\n",err,requestPayload)
		return nil, err
	}

	request.Header.Add("Authorization", accessToken)
	request.Header.Add("content-type", "application/json")

	if subAccountName != "" {
		request.Header.Add("Account-Name", subAccountName)
	}

	requestDump, err := httputil.DumpRequest(request, true)
	if err != nil {
		LogW.Println(err)
	}
	LogI.Printf("Sending request: %s", string(requestDump))

	return http.DefaultClient.Do(request)
}

func sendHoneycombEvent(account string, event string, subAccountName string, eventData string) {
	if eventData == "" {
		eventData = "{}"
	}
	requestPayload := PostHoneycombRequest {
		Account: account,
		SubAccount: subAccountName,
		TechPartner: "AWS",
		IntegrationName: "lacework-aws-security-hub-cloudformation",
		Version: "$BUILD",
		Service: "AWS Security Hub",
		InstallMethod: "cloudformation",
		Function: "setup",
		Event: event,
		EventData: eventData,
	}
	if payloadBytes, err := json.Marshal(requestPayload); err == nil {
		if request, err := http.NewRequest(http.MethodPost, "https://api.honeycomb.io/1/events/$DATASET", bytes.NewBuffer(payloadBytes)); err == nil {
			request.Header.Add("X-Honeycomb-Team", "$HONEY_KEY")
			request.Header.Add("content-type", "application/json")
			if resp, err := http.DefaultClient.Do(request); err == nil {
				LogI.Println("Set event to Honeycomb: ", event, resp.StatusCode)
			} else {
				LogW.Println("Unable to send event to Honeycomb: ", err)
			}
		}
	}
}

