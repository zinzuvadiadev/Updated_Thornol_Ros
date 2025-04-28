# Thornol

## API

### Execute

A unix domain request to this API will publish the command to the topic `topic/execute` whose listener will execute the command on console (OS SYSCALL)

#### Topic

Topic Name: `topic/exectue`

#### Unix Request

```bash
curl -s -N --unix-socket /tmp/thornol.sock -X POST -d '{"command": "ls"}' -H "Content-Type: application/json" http://localhost/execute
{"message":"data","status":true}
```

#### Request/Result

- Reqeust
```json
{
    "command": "ls"
}
```

- Result
```json
{
    "status": true,
    "message": "ls"
}
```

### Trigger

A heartbeat check to see if thoronol mqtt works (POST DATA TO MQTT)

#### Topic

Topic Name: `topic/trigger`

#### Unix Request

```bash
curl -s -N --unix-socket /tmp/thornol.sock http://localhost/trigger/ggwp
{"message":"ggwp","status":true}
```

#### Request/Result

- Reqeust
```bash
http://localhost/trigger/:DATA
```

- Result
```json
{
    "status": true,
    "message": ":DATAg"
}
```