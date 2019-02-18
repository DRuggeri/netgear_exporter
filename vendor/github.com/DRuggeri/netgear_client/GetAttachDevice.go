package netgear_client

import (
	"encoding/xml"
	"fmt"
	"html"
	"strings"
)

func (client *NetgearClient) GetAttachDevice() ([]map[string]string, error) {
	const ACTION = "urn:NETGEAR-ROUTER:service:DeviceInfo:1#GetAttachDevice"
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
    <M1:GetAttachDevice xsi:nil="true" />
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

	response, err := client.send_request(ACTION, fmt.Sprintf(REQUEST, client.sessionid), true)
	if err != nil {
		return make([]map[string]string, 0), err
	}

	var inside Node
	err = xml.Unmarshal(response, &inside)
	if err != nil {
		return make([]map[string]string, 0), fmt.Errorf("Failed to unmarshal response from inside SOAP body: %v", err)
	}

	devices := make([]map[string]string, 0)

	/* Values are HTML encoded - this breaks the ";" splitting later */
	data := html.UnescapeString(inside.Nodes[0].Content)
	results := strings.Split(data, "@")
	for i := 1; i < len(results); i++ {
		fields := strings.Split(results[i], ";")

		infoMap := map[string]string{
			"IPAddress":              fields[1],
			"Name":                   fields[2],
			"MACAddress":             fields[3],
			"ConnectionType":         fields[4],
			"WirelessLinkSpeed":      fields[5],
			"WirelessSignalStrength": fields[6],
		}
		devices = append(devices, infoMap)
	}
	return devices, nil
}
