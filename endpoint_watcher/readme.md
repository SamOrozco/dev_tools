# Endpoint Watcher

Endpoint watcher is a cli tool to monitor anything within reach of an http request. Configure this tool using yaml to alert you when
something doesn't go as planned. 

# Configuration

The watcher **configuration** is made up three main parts. The `endpoint` which is the request you wish to make specifying the `url`, 
`method` and `body`(if applicable, you can leave it out). The `js` or the javascript you are using to evaluate the response. Finally, how to handle when your js evaluation success. 

Here the simplest example of a *watcher* configuration. 

```yaml
endpoint:
  url: 'https://google.com'
  method: 'GET'
js:
  type: 'script'
  javascript: >
    var def = statusCode === 200;
success: # if the type is omitted here it will default to 'desktop' or a desktop notification
  - message: 'Google returned a 200' 
# notice the array of success's here you can have any number of actions of a success
```


## Endpoint

The endpoint is the request you wish to send. Here are all the options possible with the Endpoint object.
You can omit the `auth` and `body` fields if you want

```yaml
endpoint: 
  url: 'http://my.site.com' # web address you wish to call
  method: 'GET' # POST, PUT, DELETE, PATCH
  body: '{"userId": 23231, "name", "some payload"}'
    auth:
      type: 'basic' # basic is the only type right now. 
      username: 'my_username'
      password: 'my_password' # NOTE you can also pass in any env var with the ${var} syntax  
```


# Js or 'javascript'

The javascript object is the part of the tool that will evaluate success from the responseBody and the statusCode
of the request defined by `endpoint`. The main rules of evaluating the javascript is that you will be supplied to variables
upon executing, they are `statusCode` (the status code of the response) and `responseBody` (the response body of the request). 
You **must** also set the var `def` to true for the watcher to pass. 

You can define javascript in two ways.

**By File** if you omit the `type` it will default to the `file` type.
```yaml
js: 
  javascript: 'C:\source\to\js\file.js'
```


**Inline Script** you must specify `type: 'script'` for this to work
```yaml
js: 
  type: 'script'
  javascript: 'var def = statusCode === 200'
```

**Multi line inline script**
```yaml
js:
  type: 'script'
  javascript: >
    var json = JSON.parse(responseBody);
    var build = json["build"];
    var build_running = build["build_running"];
    var def = build_running;
```


## Success
Success if probably the most import part of this tool. The success part will be execute when your javascript has been
evaluated successful. There are currently 4 types of success actions you can take: `desktop`(desktop notifications), `webhook` (execute a web hook), 
`js` execute a javascript object (you also have access to `responseBody` and `statusCode`), and `watcher` (you can execute another watcher when one has finished successfully.)

**Desktop**
```yaml
success: # if the type is omitted here it will default to 'desktop' or a desktop notification
  - message: 'Google returned a 200' # send desktop notification containing text
```

**WebHook** execute a webhook on success
```yaml
success: # if the type is omitted here it will default to 'desktop' or a desktop notification
  - type: 'webhook'
    endpoint:
      url: 'http://my.account.create'
      method: 'POST'
      body: '{"name", "sam"}'
```

**js** execute js on success
```yaml
success: # if the type is omitted here it will default to 'desktop' or a desktop notification
  - type: 'js'
    js:
      type: 'script'
      javascript: >
        if (statusCode == 200) {
          console.log(responseBody);
        }
```

**watcher** execute a watcher on success on the current watcher you can nest these as far as you want.
```yaml
endpoint:
  url: 'https://google.com'
  method: 'GET'
js:
  type: 'script'
  javascript: >
    var def = statusCode === 200;
success:
  - type: 'watcher'
    config:
      endpoint:
        url: 'https://google.com'
        method: 'GET'
      js:
        type: 'script'
        javascript: >
          var def = statusCode === 200;
      success:
        - message: 'finished a show'
```

### Hot Tip
You can also define a watcher config with an external file. If the `config_file` field is set, you will always try to load a file. 
```yaml
config_file: 'C:\Yaml\files.yaml'
```

```yaml
endpoint:
  url: 'https://google.com'
  method: 'GET'
js:
  type: 'script'
  javascript: >
    var def = statusCode === 200;
success:
  - type: 'watcher'
    config:
      config_file: 'C:\my\yaml\file.yaml'
```

**Put it all together**

```yaml
endpoint:
  url: 'https://google.com'
  method: 'GET'
js:
  type: 'script'
  javascript: >
    var def = statusCode === 200;
success:
  - type: 'js'
    javascript: >
      setEnv('responseBody' responseBody)
  - type: 'watcher'
    config:
      config_file: 'C:\my\yaml\file.yaml'
  - type: 'webhook'
    endpoint:
      url: 'http://google.com'
      method: 'GET'
```

There are also two functions you can use in your javascript to pass variables from watcher to watcher or step to step. 
They are a `setEnv('key', 'value')` or `getEnv('key')`



# Usage 
```
ew C:\path\to\yaml.yaml
```
