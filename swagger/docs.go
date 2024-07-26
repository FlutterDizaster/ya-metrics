// Package swagger Code generated by swaggo/swag. DO NOT EDIT
package swagger

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {
            "name": "Dmitriy Loginov",
            "email": "dmitriy@loginoff.space"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/": {
            "get": {
                "description": "Get all metrics",
                "produces": [
                    "html/json"
                ],
                "tags": [
                    "metrics"
                ],
                "summary": "Get all metrics",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/view.Metric"
                            }
                        }
                    },
                    "500": {
                        "description": "Error",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/ping": {
            "get": {
                "description": "Ping DB donnection",
                "produces": [
                    "text/plain"
                ],
                "tags": [
                    "health"
                ],
                "summary": "Ping",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "Error",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/update": {
            "post": {
                "description": "Update metric in DB in JSON format",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "metrics"
                ],
                "summary": "Update metric",
                "parameters": [
                    {
                        "description": "Metric",
                        "name": "metric",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/view.Metric"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/view.Metric"
                        }
                    },
                    "400": {
                        "description": "Bad request",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "Error",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/update/{kind}/{name}/{value}": {
            "post": {
                "description": "Update metric in DB",
                "produces": [
                    "text/plain"
                ],
                "tags": [
                    "metrics"
                ],
                "summary": "Update metric",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Metric kind",
                        "name": "kind",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "Metric name",
                        "name": "name",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "Metric value",
                        "name": "value",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "Bad request",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "Error",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/updates": {
            "post": {
                "description": "Update metrics in DB",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "metrics"
                ],
                "summary": "Update metrics",
                "parameters": [
                    {
                        "description": "Metrics",
                        "name": "metrics",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/view.Metric"
                            }
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/view.Metric"
                            }
                        }
                    },
                    "400": {
                        "description": "Bad request",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "Error",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/value": {
            "post": {
                "description": "Get metric in JSON format",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "metrics"
                ],
                "summary": "Get metric",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/view.Metric"
                        }
                    },
                    "404": {
                        "description": "Metric not found",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "Error",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/value/{kind}/{name}": {
            "get": {
                "description": "Get metric",
                "produces": [
                    "text/plain"
                ],
                "tags": [
                    "metrics"
                ],
                "summary": "Get metric",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Metric kind",
                        "name": "kind",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "Metric name",
                        "name": "name",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Metric value",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "404": {
                        "description": "Metric not found",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "Error",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "view.Metric": {
            "type": "object",
            "properties": {
                "delta": {
                    "description": "Counter value\nRequired: false",
                    "type": "integer"
                },
                "id": {
                    "description": "Metric ID\nRequired: true",
                    "type": "string"
                },
                "type": {
                    "description": "Metric Type\nPossible values: gauge, counter\nRequired: true",
                    "type": "string"
                },
                "value": {
                    "description": "Gauge value\nRequired: false",
                    "type": "number"
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "0.3",
	Host:             "",
	BasePath:         "/",
	Schemes:          []string{},
	Title:            "Ya-Metrics API",
	Description:      "API for getting and setting metrics",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}