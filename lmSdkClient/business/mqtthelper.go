package business

import "github.com/eclipse/paho.golang/paho"

func GeneProps(responseTopic, jwtToken, deviceId, businessTypeStr, businessSubTypeStr, taskIdStr, codeStr, errormsg string) *paho.PublishProperties {
	props := &paho.PublishProperties{}
	if responseTopic != "" {
		props.ResponseTopic = responseTopic
	}
	props.User = props.User.Add("jwtToken", jwtToken)
	props.User = props.User.Add("deviceId", deviceId)
	props.User = props.User.Add("businessType", businessTypeStr)
	props.User = props.User.Add("businessSubType", businessSubTypeStr)
	props.User = props.User.Add("taskId", taskIdStr)
	props.User = props.User.Add("code", codeStr)
	props.User = props.User.Add("errormsg", errormsg)

	return props
}
