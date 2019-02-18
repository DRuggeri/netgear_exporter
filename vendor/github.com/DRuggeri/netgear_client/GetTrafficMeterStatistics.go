package netgear_client

import (
	"encoding/xml"
	"fmt"
	"strings"
)

func (client *NetgearClient) GetTrafficMeterStatistics() (map[string]string, error) {
	const ACTION = "urn:NETGEAR-ROUTER:service:DeviceConfig:1#GetTrafficMeterStatistics"
	const REQUEST = `<?xml version="1.0" encoding="UTF-8" standalone="no"?>
<SOAP-ENV:Envelope
  xmlns:SOAPSDK1="http://www.w3.org/2001/XMLSchema"
  xmlns:SOAPSDK2="http://www.w3.org/2001/XMLSchema-instance"
  xmlns:SOAPSDK3="http://schemas.xmlsoap.org/soap/encoding/"
  xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/">
  <SOAP-ENV:Header>
    <SessionID>%s</SessionID>
  </SOAP-ENV:Header>
  <SOAP-ENV:Body>
    <M1:GetTrafficMeterStatistics xmlns:M1="urn:NETGEAR-ROUTER:service:DeviceConfig:1">
    </M1:GetTrafficMeterStatistics>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

	response, err := client.send_request(ACTION, fmt.Sprintf(REQUEST, client.sessionid), true)
	if err != nil {
		return make(map[string]string), err
	}

	var inside Node
	err = xml.Unmarshal(response, &inside)
	if err != nil {
		return make(map[string]string), fmt.Errorf("Failed to unmarshal response from inside SOAP body: %v", err)
	}

	var stats = make(map[string]string)
	for _, node := range inside.Nodes {
		name := node.XMLName.Local
		value := strings.Replace(node.Content, ",", "", -1)
		if strings.HasPrefix(name, "New") {
			name = name[3:]
		}

		idx := strings.Index(value, "/")
		if idx > 0 {
			stats[name+"Average"] = value[idx+1:]
			stats[name] = value[0:idx]
		} else {
			stats[name] = value
		}
	}
	return stats, nil
}
