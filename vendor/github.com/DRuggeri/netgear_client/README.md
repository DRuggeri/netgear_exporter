netgear_client
==============

Similar to [pyNetgear](https://github.com/MatMaul/pynetgear), `netgear_client` provides a Go implementation of a SOAP client for communicating with Netgear routers. Unfortunately, Netgear does not release specifications for their SOAP API, so this is primarily built by reverse engineering calls.


### client
The client will cache its authentication and will attempt to reauthenticate only if a call fails due to a 401 SOAP response code. Causes for such a response code are not being authenticated (yet) and expiration of the session token. If a call is made and the authentication token cannot be obtained, an error will be propagated during the call.

To create a client:
```
client, err := NewNetgearClient(url, insecure, username, password, timeout, debug)
```
Parameters:
- **url** - The URL of the router. If this is an empty string, it defaults to 'https://routerlogin.net'. If you are on the network and using the router for DNS, several URL's are valid. Both `http://` and `https://` schemes have been observed to work for (www.)routerlogin.(net|com). Some products answer on port `5000`. Orbi devices should also support (www.)orbilogin.(net|com). This URL should not have a trailing slash.
 - **insecure** - If the URL scheme is `https://`, this disables certificate verification. This is primarily useful if you are not using the router for DNS and instead have set the `url` parameter to an IP address. Because the certificate presented by the router does not present IP SAN entries, validation will fail.
- **username** - The user name to use when authenticating with the router. If this is an empty string, it defaults to 'admin'.
- **password** - The password to use when authenticating with the router. This must be set.
- **timeout** - The number of seconds to use for HTTP timeout. Given that this client is primarily used for LANs, the timeout value should be set low (2) to propagate connection errors faster.
- **debug** - Whether to enable debug of the client itself. When set to `true`, all request and response data will be logged to STDOUT. This *contains sensitive information*, but can help identify errors easier.

Return:
- **client** - The client struct for you to make calls (listed below) with.
- **err** - Will be `nil` on success. Any errors encountered while attempting to create the client will be returned.

### Actions
The following actions are available to a client

---
#### `client.LogIn()`
This will cause the client to initiate the login flow. It is not neccessary to call `LogIn` before other calls because if any action detects that client is not already logged in, the `LogIn` flow will automatically be handled.

Usage:
`error = client.LogIn()`

Return:
- **err** - Will be `nil` on success.

---
#### `client.GetSystemInfo()`
Obtains a map of k/v pairs with system statistics. The keys and values come exactly from the API with the following exceptions:
- For newer APIs, the "New" prefix is removed from the keys

**NOTE:** CPU utilization always seems to be 100%. This is suspected to be because gathering the statistics causes CPU strain.

Usage:
`response, error = client.GetSystemInfo()`

Return:
- **response** - *`map[string]string`* - Key/value pairs straight from the API
- **err** - Will be `nil` on success.

The current keys returned, and their format is:
```
AvailableFlash => unknown (float)
CPUUtilization => percentage (int, see note)
PhysicalMemory => megabytes (int)
MemoryUtilization => percentage (int)
PhysicalFlash => megabytes (int)
```

---
#### `client.GetTrafficMeterStatistics()`
Obtains a map of k/v pairs with traffic statistics. The keys and values come exactly from the API with the following exceptions:
- For newer APIs, the "New" prefix is removed from the keys
- Some API values contain both current and average values in the form of `current/avg`. These are split into two separate keys

Usage:
`response, error = client.LogIn()`

Return:
- **response** - *`map[string]string`* - Key/value pairs straight from the API
- **err** - Will be `nil` on success.

The current keys returned, and their format is:
```
TodayConnectionTime => hh:mm
TodayDownload => megabytes (int?)
TodayUpload => megabytes (float)
YesterdayConnectionTime => hh:mm
YesterdayDownload => megabytes (int?)
YesterdayUpload => megabytes (float)
WeekConnectionTime => hh:mm
WeekDownload => megabytes (int?)
WeekDownloadAverage => megabytes (int?)
WeekUpload => megabytes (float)
WeekUploadAverage => megabytes (float)
MonthConnectionTime => hh:mm
MonthDownload => megabytes (int?)
MonthDownloadAverage => megabytes (int?)
MonthUpload => megabytes (float)
MonthUploadAverage => megabytes (float)
LastMonthConnectionTime => hh:mm
LastMonthDownload => megabytes (int?)
LastMonthDownloadAverage => megabytes (int?)
LastMonthUpload => megabytes (float)
LastMonthUploadAverage => megabytes (float)
```

---
#### `client.GetAttachDevice()`
Obtains an array of maps of k/v pairs with client information.

Usage:
`response, error = client.GetAttachDevice()`

Return:
- **response** - *`[]map[string]string`* - An array of maps containing data
- **err** - Will be `nil` on success.

The current keys returned, and their format is:
```
IPAddress => address (string)
Name => name (string)
MACAddress => address (string)
ConnectionType => wired|5G|2.4G (string)
WirelessLinkSpeed => megabits/sec (int)
WirelessSignalStrength => percentage (int)
```

---
#### `client.GetAttachDevice2()`
Obtains an array of maps of k/v pairs with client information exactly as-is from the API. While this call returns more data, it also takes longer indicating it may be more stressful on the router.

Usage:
`response, error = client.GetAttachDevice2()`

Return:
- **response** - *`[]map[string]string`* - An array of maps with key/value pairs straight from the API
- **err** - Will be `nil` on success.

The current keys returned, and their format is:
```
IP => address (string)
Name => name (string)
NameUserSet => true|false (boolean)
MAC => address (string)
ConnectionType => wired|5G|2.4G (string)
SSID => name (string, sometimes empty)
Linkspeed => megabits/sec (int, sometimes empty)
SignalStrength => percentage (int, always 100 for wired)
AllowOrBlock => Allow|Block (string)
Schedule => true|false (boolean)
DeviceType => internal id (int)
DeviceTypeUserSet => true|false (boolean)
Upload => unknown (float, apparently always 0.00)
Download => unknown (float, apparently always 0.00)
QosPriority => priority (int)
```
