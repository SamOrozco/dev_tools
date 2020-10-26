# Endpoint watcher

Simple configurable endpoint watcher. I have found this useful when waiting for builds to finish I just configure an endpoint watcher and can 
passively monitor build in background.
Configure yaml file to set up the endpoint you want to watch, then write a js file to evaluate the http response.


## yaml
The yaml is pretty self-explanatory . endpoint you wish to call, js_file you wish to evaluate http response, 
limit the number of attempts, interval_millis the number of millis to wait between requests, success_message the message
to display in form of desktop notification to user.
```yaml
endpoint:
  url: 'http://localhost:9005/test'
  method: 'GET'
js_file: 'C:\Users\samue\projects\dev_tools\endpoint_watcher\examples\template_json.js'
limit: 100
interval_millis: 1000
success_message: 'user equals sam'
```


## Js
The js is pretty simple there are only three rules. one the status code comes in the vm with the variable name `statusCode`
and the response body comes in to the vm as a string `responseBody`. Finally you must set a `def` value in the vm before exiting. 
See below: 

```javascript
var code = statusCode;
var body = responseBody;
var userObject = JSON.parse(body);
var def = userObject["user_name"] === "sam";
```

