## description of special parameters

| key      | value                       | description                                                                                                                                                                                          |
|----------|-----------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| static   | boolean                     | if the data is static,then it has no extra `get`                                                                                                                                                     |
| type     | `merge`/`route`/`get`/`api` | - `merge`:merge two or more data into a single one<br/>- `route`:build a route to call other data<br/>- `get`:get data from the server<br/>- `get`:call `[repo]`<br/>- `api`:call an api to get data |
| language | `natural`/`machine`         | - `natural`:using words containing `if/then`<br/>- `machine`:using `[input]`and`[output]`                                                                                                            

