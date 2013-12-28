# Web API — Facette

## General Information

This documentation describes RESTful web API coming along with [Facette][0] web front-end.

Unless otherwise stated, all request and response bodies use `application/json` as `Content-Type`.

The dates returned by the API are using the [RFC 3399][1] timestamp format.

## API

### Catalog

#### Origins

##### List origins

```
GET /catalog/origins
```

Returns an array listing the available origins.

Optional parameters:

 * __filter:__ the [match][3] pattern to apply on origins names (type: `string`)
 * __limit:__ the maximum number of items to return (type: `integer`)
 * __offset:__ the offset to start fetching from (type: `integer`)

Response:

```javascript
[
    "origin0",
    "origin1"
]
```

A `X-Total-Records` HTTP header containing the total number of records is returned along with the response.

##### Get a single origin

```
GET /catalog/origins/<name>
```

Returns an origin object along with the date of its last update.

Response:

```javascript
{
    "name": "origin0",
    "backend": "rrd",
    "updated": "2013-01-02T12:34:56+01:00"
}
```

#### Sources

##### List sources

```
GET /catalog/sources
```

Returns an array listing the available sources.

Optional parameters:

 * __filter:__ the [match][3] pattern to apply on sources names (type: `string`)
 * __limit:__ the maximum number of items to return (type: `integerr`)
 * __offset:__ the offset to start fetching from (type: `integer`)
 * __origin:__ the identifier of the origin to filter on (type: `string`)

Response:

```javascript
[
    "source0",
    "source1"
]
```

A `X-Total-Records` HTTP header containing the total number of records is returned along with the response.

##### Get a single source

```
GET /catalog/sources/<name>
```

Returns a source object along with the date of its last update and the list of the associated origins.

Response:

```javascript
{
    "name": "source0",
    "origins": [
        "origin0"
    ],
    "updated": "2013-01-02T12:34:56+01:00"
}
```

#### Metrics

##### List metrics

```
GET /catalog/metrics
```

Returns an array listing the available metrics.

Optional parameters:

 * __filter:__ the [match][3] pattern to apply on metrics names (type: `string`)
 * __limit:__ the maximum number of items to return (type: `integer`)
 * __offset:__ the offset to start fetching from (type: `integer`)
 * __origin:__ the identifier of the origin to filter on (type: `string`)
 * __source:__ the identifier of the source to filter on (type: `string`)

Response:

```javascript
[
    "metric0",
    "metric1"
]
```

A `X-Total-Records` HTTP header containing the total number of records is returned along with the response.

##### Get a single metric

```
GET /catalog/metrics/<name>
```

Returns a metric object along with the date of its last update and the list of the associated origins and sources.

Response:

```javascript
{
    "name": "metric0",
    "origins": [
        "origin0"
    ],
    "sources": [
        "source0",
        "source1"
    ],
    "updated": "2013-01-02T12:34:56+01:00"
}
```

##### Expand query tuples

```
POST /catalog/expand
```

Takes an array of query tuples from the request body and returns their expanded versions.

A query tuple consists of a list of 3 values: origin, source and metric.

When querying the back-end, both source and metric fields can be a reference to a group, then beginning with `group:`
prefix.

```javascript
[
        ["origin0", "host1.renm.org", "group:group0"]
]
```

Response:

```javascript
[
    [
        ["origin0", "host1.renm.org", "metric0"],
        ["origin0", "host1.renm.org", "metric1"],
        ["origin0", "host1.renm.org", "metric2"],
        ["origin0", "host1.renm.org", "metric3"]
    ]
]
```

### Library

All library items are identified by an [universally unique identifier][4] (UUID), each 36 characters long.

#### Groups

##### List groups

```
GET /library/sourcegroups
GET /library/metricgroups
```

Returns an array of objects listing the available groups.

Optional parameters:

 * __filter:__ the [match][3] pattern to apply on group names (type: `string`)
 * __limit:__ the maximum number of items to return (type: `integer`)
 * __offset:__ the offset to start fetching from (type: `integer`)

Response:

```javascript
[
    {
        "id": "386c8361-517f-404e-6c34-870983ab66e8",
        "name": "group0",
        "description": "A great group description.",
        "modified": "2013-01-02T12:34:56+01:00"
    }
]
```

A `X-Total-Records` HTTP header containing the total number of records is returned along with the response.

##### Get a single group

```
GET /library/sourcegroups/<id>
GET /library/metricgroups/<id>
```

Returns a group object along with its name, description and matching rules entries.

Response:

```javascript
{
    "id": "386c8361-517f-404e-6c34-870983ab66e8",
    "name": "group0",
    "description": "A great group description.",
    "entries": [
        {
            "origin": "origin0",
            "pattern": "glob:example.*"
        }
    ]
}
```

##### Create a new group

```
POST /library/sourcegroups
POST /library/metricgroups
```

Takes a group from the request body and stores it in the library, then returns a `Location` HTTP header pointing to the
newly created item location.

Optional parameters:

 * __inherit:__ the UUID of the group item to inherit from (type: `string`)

Possible status codes:

 * __201 Created:__ the group item has been successfully created
 * __404 Not Found:__ the group item to inherit from does not exist
 * __409 Conflict:__ another group with the same name already exists

See _Get a single group_ above for group object format.

##### Update an existing group

```
PUT /library/sourcegroups/<id>
PUT /library/metricgroups/<id>
```

Takes a group from the request body and overwrites an existing library group item.

Possible status codes:

 * __404 Not Found:__ the item to overwrite does not exist
 * __409 Conflict:__ another group with the same name already exists

See _Get a single group_ above for group object format.

##### Delete an existing group

```
DELETE /library/sourcegroups/<id>
DELETE /library/metricgroups/<id>
```

Removes an existing group item from the library.

Possible status codes:

 * __404 Not Found:__ the item to delete does not exist

#### Graphs

##### List graphs

```
GET /library/graphs
```

Returns an array of objects listing the available graphs.

Optional parameters:

 * __collection:__ the identifier of the collection to filter on (type: `string`)
 * __filter:__ the [match][3] pattern to apply on graph names (type: `string`)
 * __limit:__ the maximum number of items to return (type: `integer`)
 * __offset:__ the offset to start fetching from (type: `integer`)

Response:

```javascript
[
    {
        "id": "909fe2df-3064-4ee2-5f52-4eca2c953c76",
        "name": "graph0",
        "description": "A great graph description.",
        "modified": "2013-01-02T12:34:56+01:00"
    }
]
```

A `X-Total-Records` HTTP header containing the total number of records is returned along with the response.

##### Get a single graph

```
GET /library/graphs/<id>
```

Returns a graph object along with its name, description, type and stacks definitions.

Response:

```javascript
{
    "id": "909fe2df-3064-4ee2-5f52-4eca2c953c76",
    "name": "graph0",
    "description": "A great graph description.",
    "type": 1,
    "stacks": [
        {
            "groups": [
                {
                    "series": [
                        {
                            "metric": "metric0",
                            "source": "source0",
                            "origin": "origin0",
                            "name": "serie0"
                        }
                    ],
                    "type": 0,
                    "name": "serie0"
                }
            ],
            "name": "stack0"
        }
    ],
    "stack_mode": 0
}
```

##### Create a new graph

```
POST /library/graphs
```

Takes a graph from the request body and stores it in the library, then returns a `Location` HTTP header pointing to the
newly created item location.

Optional parameters:

 * __inherit:__ the UUID of the graph item to inherit from (type: `string`)

Possible status codes:

 * __201 Created:__ the graph item has been successfully created
 * __404 Not Found:__ the graph item to inherit from does not exist
 * __409 Conflict:__ another graph with the same name already exists

See _Get a single graph_ above for graph object format.

##### Update an existing graph

```
PUT /library/graphs/<id>
```

Takes a graph from the request body and overwrites an existing library graph item.

Possible status codes:

 * __404 Not Found:__ the item to overwrite does not exist
 * __409 Conflict:__ another group with the same name already exists

See _Get a single graph_ above for graph object format.

##### Delete an existing graph

```
DELETE /library/graphs/<id>
```

Removes an existing graph item from the library.

Possible status codes:

 * __404 Not Found:__ the item to delete does not exist

##### Get graphs plots values

```
POST /library/graphs/plots
```

Takes a plot request from the request body and returns graph series plots values and information.

Possible status codes:

 * __404 Not Found:__ the requested graph template does not exist

Request:

```javascript
{
    "graph": "909fe2df-3064-4ee2-5f52-4eca2c953c76",
    "range": "-1d"
}
```

Response (plots values are truncated):

```javascript
{
    "id": "909fe2df-3064-4ee2-5f52-4eca2c953c76",
    "name": "Chart name",
    "description": "A great chart description.",
    "type": 1,
    "stacks": [
        {
            "series": [
                {
                    "info": {
                        "min": 0,
                        "max": 1.023,
                        "last": 0.381,
                        "avg": 0.109164
                    },
                    "plots": [
                        0.348,
                        0.351,
                        0.42300000000000004,
                        ...
                        0,
                        0.10500000000000001,
                        0.42400000000000004,
                        0.381,
                        null
                    ],
                    "name": "serie0"
                }
            ],
            "name": "stack0"
        }
    ],
    "stack_mode": 0,
    "start": "2013-01-01T12:34:56+01:00",
    "end": "2013-01-02T12:34:56+01:00",
    "step": 9.944751381,
    "modified": "2013-01-02T12:34:56+01:00"
}
```

##### Get graphs series values

```
POST /library/graphs/values
```

Takes a plot request from the request body and returns graph series information.

Possible status codes:

 * __404 Not Found:__ the requested graph template does not exist

Request:

```javascript
{
    "graph": "909fe2df-3064-4ee2-5f52-4eca2c953c76",
    "range": "-1d"
}
```

Response:

```javascript
{
    "serie0": {
        "min": 0.02,
        "avg": 0.0315,
        "max": 0.043,
        "last": 0.02
    }
}
```

#### Collections

##### List collections

```
GET /library/collections
```

Returns an array of objects listing the available collections.

Optional parameters:

 * __exclude:__ the identifier of the collection to exclude from listing (type: `string`)
 * __filter:__ the [match][3] pattern to apply on collection names (type: `string`)
 * __limit:__ the maximum number of items to return (type: `integer`)
 * __offset:__ the offset to start fetching from (type: `integer`)
 * __parent:__ the identifier of the parent collection to filter on (type: `string`)

Response:

```javascript
[
    {
        "id": "916f955a-752c-468f-61b3-ace173a6d2da",
        "name": "collection0",
        "description": "A great collection description.",
        "parent": null,
        "has_children": false,
        "modified": "2013-01-02T12:34:56+01:00"
    }
]
```

A `X-Total-Records` HTTP header containing the total number of records is returned along with the response.

##### Get a single collection

```
GET /library/collections/<id>
```

Returns a collection object along with its name, description and graphs entries.

Response:

```javascript
{
    "id": "916f955a-752c-468f-61b3-ace173a6d2da",
    "name": "collection0",
    "description": "A great collection description.",
    "entries": [
        {
            "options": {
                "title": "Chart title",
                "sample": "400",
                "range": "-1mo",
                "percentiles": "95",
                "constants": ""
            },
            "id": "909fe2df-3064-4ee2-5f52-4eca2c953c76"
        }
    ]
}
```

##### Create a new collection

```
POST /library/collections
```

Takes a collection from the request body and stores it in the library, then returns a `Location` HTTP header pointing to the
newly created item location.

Optional parameters:

 * __inherit:__ the UUID of the collection item to inherit from (type: `string`)

Possible status codes:

 * __201 Created:__ the collection item has been successfully created
 * __404 Not Found:__ the collection item to inherit from does not exist
 * __409 Conflict:__ another collection with the same name already exists

See _Get a single collection_ above for group object format.

##### Update an existing collection

```
PUT /library/collections/<id>
```

Takes a collection from the request body and overwrites an existing library collection item.

Possible status codes:

 * __404 Not Found:__ the item to overwrite does not exist
 * __409 Conflict:__ another group with the same name already exists

See _Get a single collection_ above for group object format.

##### Delete an existing collection

```
DELETE /library/collections/<id>
```

Removes an existing collection item from the library.

Possible status codes:

 * __404 Not Found:__ the item to delete does not exist

### Main

#### Get items statistics

```
GET /stats
```

Returns a statistics object along with back-end items information.

Response:

```javascript
{
    "catalog_updated": "2013-01-02T12:34:56+01:00",
    "groups": 1,
    "collections": 1,
    "graphs": 1,
    "metrics": 353,
    "sources": 3,
    "origins": 1
}
```

#### Reload service configuration

```
GET /reload
```

Reloads the service configuration files and refreshes back-ends.

Possible status codes:

 * __503 Service Unavailable:__ the server is reloading its configuration

### HTTP Status Codes

#### 200 OK

The request was successful and the server is returning data in the response body.

#### 201 Created

The request was successful resulting in the creation of a new item. The response body is empty.

#### 400 Bad Request

The request or the data supplied along with the request is invalid, could not be processed or when mandatory fields are
not met.

#### 401 Unauthorized

The authentication credentials are missing or invalid.

#### 404 Not Found

The requested item or an item to inherit from could not be found. Could be returned by `POST` creation requests.

#### 405 Method Not Allowed

The requested method is not supported the API resource.

#### 409 Conflict

An item with the same name and type already exists. Could be returned by `POST` and `PUT` creation requests.

#### 415 Unsupported Media Type

The request used a `Content-Type` not supported by the API resource.

#### 503 Service Unavailable

The server is currently unavailable. Such response is sent when the server is being initialized or reloading its
configuration.


[0]: http://facette.io/
[1]: http://www.ietf.org/rfc/rfc3339.txt
[3]: http://golang.org/pkg/path/#Match
[4]: http://en.wikipedia.org/wiki/Universally_unique_identifier
