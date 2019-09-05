# cleta

**cleta** = **cl**oud m**eta**data 

|ENV_VAR|Type||
|-|-|-|-|
|CLETA_METADATA_BIND_ADDR|HostPort|
|CLETA_METADATA_STORE|enum|`dir`\|`postgres`\|`mariadb`
|CLETA_METADATA_STORE_DIR|path|`a:b:c`

|Flag|Type|Multiplicity||
|-|-|-|-|
|`metadata-bind-addr`|Host|many|
|`metadata-port`|Port|many|
|
|`metadata-store`|enum|once|`dir`\|`postgres`
|`metadata-store-dir`|string|many|
|`metadata-store-dir-cache-size`|int|once|
|`metadata-store-postgres`|string|once|
|
|`api-bind-addr`|IPAddr|once|
|
|`neighbor-table-refresh-interval`|time.Duration|once|1ms


```bash
cleta serve \
    --metadata-bind-addr="169.254.169.254:80" \
    --metadata-store="dir" \
    --metadata-store-dir="." \
    --metadata-store-dir="files" \
    --api-bind-addr="0.0.0.0:443"
```